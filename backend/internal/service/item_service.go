package service

import (
	"context"
	"errors"
	"time"

	"auction-platform/internal/model"
	"auction-platform/internal/repository"
)

var (
	ErrItemNotFound = errors.New("item not found")
	ErrNotYourItem  = errors.New("this item does not belong to you")
)

type ItemService struct {
	itemRepo *repository.ItemRepository
	cache    *repository.CacheRepository
}

func NewItemService(itemRepo *repository.ItemRepository, cache *repository.CacheRepository) *ItemService {
	return &ItemService{itemRepo: itemRepo, cache: cache}
}

func (s *ItemService) Create(ctx context.Context, req *CreateItemRequest, sellerID int64) (*model.Item, error) {
	now := time.Now()
	item := &model.Item{
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     req.ImageUrl,
		StartPrice:   req.StartPrice,
		CurrentPrice: req.StartPrice,
		ReservePrice: req.ReservePrice,
		BidIncrement: req.BidIncrement,
		SellerID:     sellerID,
		Status:       string(model.ItemStatusDraft),
		StartTime:    time.Unix(req.StartTime, 0),
		EndTime:      time.Unix(req.EndTime, 0),
	}

	if item.EndTime.Before(now) {
		return nil, errors.New("end time must be in the future")
	}
	if item.StartTime.After(item.EndTime) {
		return nil, errors.New("start time must be before end time")
	}
	if item.BidIncrement <= 0 {
		item.BidIncrement = 100 // 默认加价幅度
	}

	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	item, err := s.itemRepo.GetByID(id)
	if err != nil {
		return nil, ErrItemNotFound
	}
	return item, nil
}

func (s *ItemService) List(ctx context.Context, status string, sellerID int64, keyword string, page, pageSize int) ([]*model.Item, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.itemRepo.List(status, sellerID, keyword, page, pageSize)
}

func (s *ItemService) ListMyItems(ctx context.Context, userID int64, status string) ([]*model.Item, int, error) {
	return s.itemRepo.List(status, userID, "", 1, 100)
}

func (s *ItemService) Update(ctx context.Context, req *UpdateItemRequest, userID int64) (*model.Item, error) {
	item, err := s.itemRepo.GetByID(req.Id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	if item.SellerID != userID {
		return nil, ErrNotYourItem
	}

	if item.Status != string(model.ItemStatusDraft) {
		return nil, errors.New("only draft items can be updated")
	}

	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.ImageUrl != "" {
		item.ImageUrl = req.ImageUrl
	}
	if req.ReservePrice > 0 {
		item.ReservePrice = req.ReservePrice
	}
	if req.BidIncrement > 0 {
		item.BidIncrement = req.BidIncrement
	}
	if req.EndTime > 0 {
		item.EndTime = time.Unix(req.EndTime, 0)
	}

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) Publish(ctx context.Context, id int64, userID int64) (*model.Item, error) {
	item, err := s.itemRepo.GetByID(id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	if item.SellerID != userID {
		return nil, ErrNotYourItem
	}

	if item.Status != string(model.ItemStatusDraft) {
		return nil, errors.New("only draft items can be published")
	}

	item.Status = string(model.ItemStatusListed)
	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) Cancel(ctx context.Context, id int64, userID int64) (*model.Item, error) {
	item, err := s.itemRepo.GetByID(id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	if item.SellerID != userID {
		return nil, ErrNotYourItem
	}

	if item.Status == string(model.ItemStatusSold) || item.Status == string(model.ItemStatusCancelled) {
		return nil, errors.New("cannot cancel this item")
	}

	item.Status = string(model.ItemStatusCancel)
	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) UpdatePrice(ctx context.Context, id, price int64) error {
	return s.itemRepo.UpdatePrice(id, price)
}

// DTOs
type CreateItemRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageUrl     string `json:"image_url"`
	StartPrice   int64  `json:"start_price"`
	ReservePrice int64  `json:"reserve_price"`
	BidIncrement int64  `json:"bid_increment"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
}

type UpdateItemRequest struct {
	Id           int64  `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageUrl     string `json:"image_url"`
	ReservePrice int64  `json:"reserve_price"`
	BidIncrement int64  `json:"bid_increment"`
	EndTime      int64  `json:"end_time"`
}
