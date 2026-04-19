package service

import (
	"context"
	"errors"
	"time"

	"auction-platform/internal/model"
)

var (
	ErrBidTooLow      = errors.New("bid amount is too low")
	ErrAuctionEnded    = errors.New("auction has ended")
	ErrCannotBidOwn   = errors.New("cannot bid on your own item")
	ErrItemNotActive  = errors.New("item is not available for bidding")
)

type BidService struct {
	bidRepo  BidRepo
	itemRepo ItemRepo
	cache    CacheRepo
}

func NewBidService(bidRepo BidRepo, itemRepo ItemRepo, cache CacheRepo) *BidService {
	return &BidService{bidRepo: bidRepo, itemRepo: itemRepo, cache: cache}
}

func (s *BidService) PlaceBid(ctx context.Context, itemID, buyerID, amount int64) (*model.Bid, int64, bool, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, 0, false, ErrItemNotFound
	}

	// 验证
	if item.SellerID == buyerID {
		return nil, 0, false, ErrCannotBidOwn
	}
	if item.Status != string(model.ItemStatusListed) {
		return nil, 0, false, ErrItemNotActive
	}
	now := time.Now()
	if now.After(item.EndTime) {
		return nil, 0, false, ErrAuctionEnded
	}
	if now.Before(item.StartTime) {
		return nil, 0, false, errors.New("auction has not started yet")
	}

	// 检查最低出价
	minBid := item.CurrentPrice + item.BidIncrement
	if amount < minBid {
		return nil, minBid, false, ErrBidTooLow
	}

	// 创建出价
	bid := &model.Bid{
		ItemID:  itemID,
		BuyerID: buyerID,
		Amount:  amount,
		Status:  string(model.BidStatusActive),
	}

	if err := s.bidRepo.Create(bid); err != nil {
		return nil, 0, false, err
	}

	// 更新商品当前价
	if err := s.itemRepo.UpdatePrice(itemID, amount); err != nil {
		return nil, 0, false, err
	}

	// 将之前的领先出价标记为 outbid
	highestBid, err := s.bidRepo.GetHighestBid(itemID)
	if err == nil && highestBid != nil && highestBid.ID != bid.ID {
		s.bidRepo.UpdateStatus(highestBid.ID, string(model.BidStatusOutbid))
	}
	bid.Status = string(model.BidStatusWinning)

	return bid, amount, true, nil
}

func (s *BidService) GetBidsByItemID(ctx context.Context, itemID int64) ([]*model.Bid, int64, int64, error) {
	bids, err := s.bidRepo.ListByItemID(itemID)
	if err != nil {
		return nil, 0, 0, err
	}

	var highestPrice, highestBidder int64
	if len(bids) > 0 {
		highestPrice = bids[0].Amount
		highestBidder = bids[0].BuyerID
	}

	return bids, highestPrice, highestBidder, nil
}

func (s *BidService) GetMyBids(ctx context.Context, buyerID int64) ([]*model.Bid, error) {
	return s.bidRepo.ListByBuyerID(buyerID)
}

func (s *BidService) CountBidsByItemID(ctx context.Context, itemID int64) (int64, error) {
	return s.bidRepo.CountByItemID(itemID)
}
