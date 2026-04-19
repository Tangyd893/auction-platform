package service

import (
	"context"
	"fmt"
	"io"
	"sync"

	"auction-platform/proto/gen/auction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuctionService struct {
	auction.UnimplementedAuctionServiceServer
	authService  *AuthService
	itemService  *ItemService
	bidService   *BidService
	orderService *OrderService
	userService  *UserService

	// 订阅管理
	auctionSubs   map[int64][]chan *auction.AuctionUpdate
	subMutex      sync.RWMutex
}

func NewAuctionService(auth *AuthService, item *ItemService, bid *BidService, order *OrderService, user *UserService) *AuctionService {
	svc := &AuctionService{
		authService:   auth,
		itemService:   item,
		bidService:    bid,
		orderService:  order,
		userService:   user,
		auctionSubs:   make(map[int64][]chan *auction.AuctionUpdate),
	}
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
	serviceReq := &CreateItemRequest{
		Title:        req.Title,
		Description:  req.Description,
		ImageUrl:     req.ImageUrl,
		StartPrice:   req.StartPrice,
		ReservePrice: req.ReservePrice,
		BidIncrement: req.BidIncrement,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}
	item, err := s.itemService.Create(ctx, serviceReq, userID)
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
	bid, currentPrice, isWinning, err := s.bidService.PlaceBid(ctx, req.ItemId, userID, req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
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

// Client Streaming: 批量出价
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

		_, currentPrice, isWinning, err := s.bidService.PlaceBid(stream.Context(), req.ItemId, userID, req.Amount)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		lastBid = &auction.PlaceBidResponse{
			CurrentPrice: currentPrice,
			IsWinning:    isWinning,
			Message:      "bid received",
		}
	}
}

// Server Streaming: 订阅某个拍品的出价更新
func (s *AuctionService) StreamBids(req *auction.StreamBidsRequest, stream auction.AuctionService_StreamBidsServer) error {
	itemID := req.ItemId
	ctx := stream.Context()

	_, highestPrice, highestBidder, err := s.bidService.GetBidsByItemID(ctx, itemID)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// 发送初始状态
	if err := stream.Send(&auction.StreamBidsResponse{
		CurrentPrice: highestPrice,
		TotalBids:    0,
	}); err != nil {
		return err
	}

	_ = highestBidder

	// TODO: 实现 Redis Pub/Sub 实时推送
	<-ctx.Done()
	return ctx.Err()
}

// Bidirectional Streaming: 实时竞价窗口
func (s *AuctionService) BidirectionalBid(stream auction.AuctionService_BidirectionalBidServer) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			if bidReq := msg.GetBid(); bidReq != nil {
				userID := getUserIDFromContext(ctx)
				bid, currentPrice, isWinning, err := s.bidService.PlaceBid(ctx, bidReq.ItemId, userID, bidReq.Amount)
				if err != nil {
					stream.Send(&auction.BidStreamResponse{
						Payload: &auction.BidStreamResponse_Error{
							Error: &auction.Error{Code: 400, Message: err.Error()},
						},
					})
					continue
				}
				stream.Send(&auction.BidStreamResponse{
					Payload: &auction.BidStreamResponse_BidResult{
						BidResult: &auction.PlaceBidResponse{
							BidId:        bid.ID,
							CurrentPrice: currentPrice,
							IsWinning:    isWinning,
						},
					},
				})
			}
		}
	}
}

// Server Streaming: 拍卖大厅
func (s *AuctionService) StreamAuction(req *auction.StreamAuctionRequest, stream auction.AuctionService_StreamAuctionServer) error {
	ctx := stream.Context()

	// TODO: 实现定期推送所有订阅拍品的当前状态
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stream.Context().Done():
			return stream.Context().Err()
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
