package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"auction-platform/internal/model"
	"auction-platform/proto/gen/auction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type bidResult struct {
	resp *auction.BidStreamResponse
	err  error
}

// ============ Prometheus Metrics ============

var (
	bidiStreamsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bidirectional_streams_active",
		Help: "Currently active bidirectional streams",
	})

	bidsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bids_total",
		Help: "Total number of bids received",
	}, []string{"result"}) // result=success|rejected|error

	bidLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "bid_latency_seconds",
		Help:    "Bid processing latency",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"result"})

	activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "auction_active_connections",
		Help: "Number of active auction connections",
	})

	circuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "circuit_breaker_state",
		Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
	}, []string{"name"})
)

const (
	maxBidiStreams       = 500                  // 全局最大双向流数
	maxBidsPerSecond     = 5                    // 每用户每秒最大出价次数
	bidCircuitThreshold  = 20                  // 熔断器阈值
	bidCircuitTimeout    = 30 * time.Second    // 熔断恢复时间
	streamTimeout        = 10 * time.Minute   // 单个流最大存活时间
)

// ============ AuctionService ============

type AuctionService struct {
	auction.UnimplementedAuctionServiceServer
	authService  *AuthService
	itemService  *ItemService
	bidService   *BidService
	orderService *OrderService
	userService  *UserService

	connMgr      *ConnectionManager
	bidBreaker   *CircuitBreaker

	auctionSubs map[int64][]chan *auction.AuctionUpdate
	subMutex    sync.RWMutex

	bidMu sync.Mutex
}

func NewAuctionService(
	auth *AuthService, item *ItemService,
	bid *BidService, order *OrderService, user *UserService,
) *AuctionService {
	svc := &AuctionService{
		authService:  auth,
		itemService:  item,
		bidService:   bid,
		orderService: order,
		userService:  user,
		connMgr:      NewConnectionManager(maxBidiStreams, maxBidsPerSecond),
		bidBreaker:   NewCircuitBreaker(bidCircuitThreshold, bidCircuitTimeout),
		auctionSubs:  make(map[int64][]chan *auction.AuctionUpdate),
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			state := svc.bidBreaker.State()
			var val float64
			switch state {
			case "closed":
				val = 0
			case "open":
				val = 1
			case "half-open":
				val = 2
			}
			circuitBreakerState.WithLabelValues("bid_breaker").Set(val)
		}
	}()

	return svc
}

func (s *AuctionService) Register(ctx context.Context, req *auction.RegisterRequest) (*auction.User, error) {
	user, err := s.authService.Register(req.Username, req.Password, req.Email, req.Role.String())
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return toProtoUser(user), nil
}

func (s *AuctionService) Login(ctx context.Context, req *auction.LoginRequest) (*auction.LoginResponse, error) {
	token, user, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &auction.LoginResponse{
		Token: token,
		User:  toProtoUser(user),
	}, nil
}

func (s *AuctionService) GetUser(ctx context.Context, req *auction.GetUserRequest) (*auction.User, error) {
	user, err := s.userService.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoUser(user), nil
}

func (s *AuctionService) ListUsers(ctx context.Context, req *auction.ListUsersRequest) (*auction.ListUsersResponse, error) {
	users, total, err := s.userService.List(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auction.ListUsersResponse{
		Users: toProtoUsers(users),
		Total: int32(total),
	}, nil
}

func (s *AuctionService) CreateItem(ctx context.Context, req *auction.CreateItemRequest) (*auction.Item, error) {
	userID := getUserIDFromContext(ctx)
	svcReq := &CreateItemRequest{
		Title:        req.Title,
		Description:  req.Description,
		ImageUrl:     req.ImageUrl,
		StartPrice:   req.StartPrice,
		ReservePrice: req.ReservePrice,
		BidIncrement: req.BidIncrement,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}
	item, err := s.itemService.Create(ctx, svcReq, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoItem(item), nil
}

func (s *AuctionService) GetItem(ctx context.Context, req *auction.GetItemRequest) (*auction.Item, error) {
	item, err := s.itemService.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoItem(item), nil
}

func (s *AuctionService) ListItems(ctx context.Context, req *auction.ListItemsRequest) (*auction.ListItemsResponse, error) {
	items, total, err := s.itemService.List(ctx, req.Status.String(), req.SellerId, req.Keyword, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auction.ListItemsResponse{
		Items: toProtoItems(items),
		Total: int32(total),
	}, nil
}

func (s *AuctionService) ListMyItems(ctx context.Context, req *auction.ListMyItemsRequest) (*auction.ListItemsResponse, error) {
	items, total, err := s.itemService.ListMyItems(ctx, req.UserId, req.Status.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auction.ListItemsResponse{
		Items: toProtoItems(items),
		Total: int32(total),
	}, nil
}

func (s *AuctionService) CancelItem(ctx context.Context, req *auction.CancelItemRequest) (*auction.Item, error) {
	userID := getUserIDFromContext(ctx)
	item, err := s.itemService.Cancel(ctx, req.Id, userID)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return toProtoItem(item), nil
}

func (s *AuctionService) PlaceBid(ctx context.Context, req *auction.PlaceBidRequest) (*auction.PlaceBidResponse, error) {
	userID := getUserIDFromContext(ctx)

	if !s.bidBreaker.Allow() {
		bidsTotal.WithLabelValues("rejected").Inc()
		return nil, status.Error(codes.Unavailable, "service temporarily unavailable")
	}

	if !s.connMgr.AllowBid(userID) {
		bidsTotal.WithLabelValues("rejected").Inc()
		return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
	}

	start := time.Now()
	bid, currentPrice, isWinning, err := s.bidService.PlaceBid(ctx, req.ItemId, userID, req.Amount)
	latency := time.Since(start).Seconds()

	if err != nil {
		s.bidBreaker.RecordFailure()
		bidsTotal.WithLabelValues("error").Inc()
		bidLatency.WithLabelValues("error").Observe(latency)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.bidBreaker.RecordSuccess()
	bidsTotal.WithLabelValues("success").Inc()
	bidLatency.WithLabelValues("success").Observe(latency)

	return &auction.PlaceBidResponse{
		BidId:        bid.ID,
		CurrentPrice: currentPrice,
		IsWinning:    isWinning,
		Message:      "bid placed successfully",
	}, nil
}

func (s *AuctionService) GetBids(ctx context.Context, req *auction.GetBidsRequest) (*auction.GetBidsResponse, error) {
	bids, highestPrice, highestBidder, err := s.bidService.GetBidsByItemID(ctx, req.ItemId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auction.GetBidsResponse{
		Bids:            toProtoBids(bids),
		HighestPrice:    highestPrice,
		HighestBidderId: highestBidder,
	}, nil
}

func (s *AuctionService) GetMyBids(ctx context.Context, req *auction.GetMyBidsRequest) (*auction.GetBidsResponse, error) {
	bids, err := s.bidService.GetMyBids(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var highestPrice, highestBidder int64
	if len(bids) > 0 {
		highestPrice = bids[0].Amount
		highestBidder = bids[0].BuyerID
	}
	return &auction.GetBidsResponse{
		Bids:            toProtoBids(bids),
		HighestPrice:    highestPrice,
		HighestBidderId: highestBidder,
	}, nil
}

func (s *AuctionService) PlaceBidBatch(stream auction.AuctionService_PlaceBidBatchServer) error {
	userID := getUserIDFromContext(stream.Context())
	var lastBid *auction.PlaceBidResponse

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(lastBid)
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if !s.bidBreaker.Allow() {
			bidsTotal.WithLabelValues("rejected").Inc()
			continue
		}
		if !s.connMgr.AllowBid(userID) {
			bidsTotal.WithLabelValues("rejected").Inc()
			continue
		}

		_, currentPrice, isWinning, err := s.bidService.PlaceBid(stream.Context(), req.ItemId, userID, req.Amount)
		if err != nil {
			bidsTotal.WithLabelValues("error").Inc()
			continue
		}

		bidsTotal.WithLabelValues("success").Inc()
		lastBid = &auction.PlaceBidResponse{
			CurrentPrice: currentPrice,
			IsWinning:    isWinning,
			Message:      "bid received",
		}
	}
}

func (s *AuctionService) StreamBids(req *auction.StreamBidsRequest, stream auction.AuctionService_StreamBidsServer) error {
	ctx := stream.Context()

	if !s.connMgr.AddStream() {
		return status.Error(codes.ResourceExhausted, "too many streams")
	}
	defer s.connMgr.RemoveStream()

	_, highestPrice, highestBidder, err := s.bidService.GetBidsByItemID(ctx, req.ItemId)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := stream.Send(&auction.StreamBidsResponse{
		CurrentPrice: highestPrice,
		TotalBids:    0,
	}); err != nil {
		return err
	}

	_ = highestBidder

	<-ctx.Done()
	return ctx.Err()
}

// ============ Bidirectional Streaming: 实时竞价 ============

func (s *AuctionService) BidirectionalBid(stream auction.AuctionService_BidirectionalBidServer) error {
	ctx := stream.Context()

	if !s.connMgr.AddStream() {
		return status.Error(codes.ResourceExhausted, "too many connections")
	}
	defer s.connMgr.RemoveStream()

	bidiStreamsActive.Inc()
	activeConnections.Inc()
	defer bidiStreamsActive.Dec()
	defer activeConnections.Dec()

	ctx, cancel := context.WithTimeout(ctx, streamTimeout)
	defer cancel()

	userID := getUserIDFromContext(ctx)

	bidCh := make(chan bidResult, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(bidCh)
				return
			default:
				msg, err := stream.Recv()
				if err == io.EOF {
					close(bidCh)
					return
				}
				if err != nil {
					bidCh <- bidResult{err: err}
					return
				}

				if subscribe := msg.GetSubscribe(); subscribe != nil {
					go s.handleBidirectionalSubscribe(ctx, stream, subscribe)
					continue
				}

				if bidReq := msg.GetBid(); bidReq != nil {
					s.handleBidirectionalBid(ctx, stream, bidReq, userID, bidCh)
				}
			}
		}
	}()

	for result := range bidCh {
		if result.err != nil {
			return result.err
		}
	}

	return nil
}

func (s *AuctionService) handleBidirectionalBid(
	ctx context.Context,
	stream auction.AuctionService_BidirectionalBidServer,
	req *auction.PlaceBidRequest,
	userID int64,
	bidCh chan bidResult,
) {
	if !s.bidBreaker.Allow() {
		bidCh <- bidResult{resp: &auction.BidStreamResponse{
			Payload: &auction.BidStreamResponse_Error{
				Error: &auction.Error{Code: 503, Message: "service unavailable, please retry later"},
			},
		}}
		return
	}

	if !s.connMgr.AllowBid(userID) {
		bidCh <- bidResult{resp: &auction.BidStreamResponse{
			Payload: &auction.BidStreamResponse_Error{
				Error: &auction.Error{Code: 429, Message: "rate limit exceeded"},
			},
		}}
		return
	}

	if req.Nonce != "" && req.Timestamp != 0 {
		if !s.connMgr.ValidNonce(req.Nonce, req.Timestamp) {
			bidCh <- bidResult{resp: &auction.BidStreamResponse{
				Payload: &auction.BidStreamResponse_Error{
					Error: &auction.Error{Code: 400, Message: "invalid or replayed nonce"},
				},
			}}
			return
		}
	}

	start := time.Now()
	bid, currentPrice, isWinning, err := s.bidService.PlaceBid(ctx, req.ItemId, userID, req.Amount)
	latency := time.Since(start).Seconds()

	s.bidMu.Lock()
	defer s.bidMu.Unlock()

	if err != nil {
		s.bidBreaker.RecordFailure()
		bidsTotal.WithLabelValues("error").Inc()
		bidLatency.WithLabelValues("error").Observe(latency)

		bidCh <- bidResult{resp: &auction.BidStreamResponse{
			Payload: &auction.BidStreamResponse_Error{
				Error: &auction.Error{Code: 400, Message: err.Error()},
			},
		}}
		return
	}

	s.bidBreaker.RecordSuccess()
	bidsTotal.WithLabelValues("success").Inc()
	bidLatency.WithLabelValues("success").Observe(latency)

	s.broadcastBidUpdate(req.ItemId, bid, currentPrice)

	bidCh <- bidResult{resp: &auction.BidStreamResponse{
		Payload: &auction.BidStreamResponse_BidResult{
			BidResult: &auction.PlaceBidResponse{
				BidId:        bid.ID,
				CurrentPrice: currentPrice,
				IsWinning:    isWinning,
				Message:      "bid accepted",
			},
		},
	}}
}

func (s *AuctionService) handleBidirectionalSubscribe(
	ctx context.Context,
	stream auction.AuctionService_BidirectionalBidServer,
	req *auction.StreamBidsRequest,
) {
	itemID := req.ItemId

	_, highestPrice, highestBidder, err := s.bidService.GetBidsByItemID(ctx, itemID)
	if err == nil {
		s.bidMu.Lock()
		stream.Send(&auction.BidStreamResponse{
			Payload: &auction.BidStreamResponse_BidUpdate{
				BidUpdate: &auction.AuctionUpdate{
					ItemId:          itemID,
					CurrentPrice:    highestPrice,
					HighestBidderId: highestBidder,
				},
			},
		})
		s.bidMu.Unlock()
	}
	_ = highestBidder
}

func (s *AuctionService) broadcastBidUpdate(itemID int64, bid *model.Bid, currentPrice int64) {
	s.subMutex.RLock()
	chans, ok := s.auctionSubs[itemID]
	s.subMutex.RUnlock()

	if !ok {
		return
	}

	update := &auction.AuctionUpdate{
		ItemId:       itemID,
		CurrentPrice: currentPrice,
	}

	for _, ch := range chans {
		select {
		case ch <- update:
		default:
		}
	}
}

// StreamAuction Server Streaming: 拍卖大厅
// 客户端订阅多个拍品，服务端定期推送所有订阅拍品的最新状态
func (s *AuctionService) StreamAuction(req *auction.StreamAuctionRequest, stream auction.AuctionService_StreamAuctionServer) error {
	ctx := stream.Context()

	if !s.connMgr.AddStream() {
		return status.Error(codes.ResourceExhausted, "too many connections")
	}
	defer s.connMgr.RemoveStream()
	bidiStreamsActive.Inc()
	activeConnections.Inc()
	defer bidiStreamsActive.Dec()
	defer activeConnections.Dec()

	itemIDs := req.GetItemIds()
	if len(itemIDs) == 0 {
		return status.Error(codes.InvalidArgument, "item_ids cannot be empty")
	}
	if len(itemIDs) > 50 {
		return status.Error(codes.InvalidArgument, "item_ids cannot exceed 50")
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// 批量查询所有订阅拍品的最新状态
			for _, itemID := range itemIDs {
				item, err := s.itemService.GetByID(ctx, itemID)
				if err != nil {
					continue // 单个查失败不影响其他
				}

				bids, _, _, err := s.bidService.GetBidsByItemID(ctx, itemID)
				if err != nil {
					continue
				}

				update := &auction.AuctionUpdate{
					ItemId:       item.ID,
					Title:        item.Title,
					CurrentPrice: item.CurrentPrice,
					BidCount:     int64(len(bids)),
					EndTime:      item.EndTime.Unix(),
				}

				if err := stream.Send(update); err != nil {
					return err
				}
			}
		}
	}
}

func (s *AuctionService) CreateOrder(ctx context.Context, req *auction.CreateOrderRequest) (*auction.Order, error) {
	userID := getUserIDFromContext(ctx)
	order, err := s.orderService.CreateOrder(ctx, req.ItemId, userID)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return toProtoOrder(order), nil
}

func (s *AuctionService) GetOrder(ctx context.Context, req *auction.GetOrderRequest) (*auction.Order, error) {
	order, err := s.orderService.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoOrder(order), nil
}

func (s *AuctionService) ListOrders(ctx context.Context, req *auction.ListOrdersRequest) (*auction.ListOrdersResponse, error) {
	orders, total, err := s.orderService.ListOrders(ctx, req.UserId, "", int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auction.ListOrdersResponse{
		Orders: toProtoOrders(orders),
		Total:  int32(total),
	}, nil
}

func (s *AuctionService) UpdateOrderStatus(ctx context.Context, req *auction.UpdateOrderStatusRequest) (*auction.Order, error) {
	order, err := s.orderService.UpdateOrderStatus(ctx, req.Id, fmt.Sprintf("%d", req.Status))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoOrder(order), nil
}
