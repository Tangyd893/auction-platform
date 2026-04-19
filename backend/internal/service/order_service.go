package service

import (
	"context"
	"errors"

	"auction-platform/internal/model"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists for this item")
)

type OrderService struct {
	orderRepo OrderRepo
	itemRepo  ItemRepo
}

func NewOrderService(orderRepo OrderRepo, itemRepo ItemRepo) *OrderService {
	return &OrderService{orderRepo: orderRepo, itemRepo: itemRepo}
}

func (s *OrderService) CreateOrder(ctx context.Context, itemID, buyerID int64) (*model.Order, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, ErrItemNotFound
	}

	// 检查是否已有订单
	existing, err := s.orderRepo.GetByItemID(itemID)
	if err == nil && existing != nil {
		return nil, ErrOrderAlreadyExists
	}

	// 检查是否有人出价
	if item.CurrentPrice <= 0 {
		return nil, errors.New("item has no bids")
	}

	// 检查成交价是否达到保留价
	if item.CurrentPrice < item.ReservePrice {
		return nil, errors.New("reserve price not met")
	}

	// 获取最高出价者
	order := &model.Order{
		ItemID:     itemID,
		SellerID:   item.SellerID,
		BuyerID:    buyerID,
		FinalPrice: item.CurrentPrice,
		Status:     string(model.OrderStatusPending),
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	// 更新拍品状态为已售
	s.itemRepo.UpdateStatus(itemID, string(model.ItemStatusSold))

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context, userID int64, status string, page, pageSize int) ([]*model.Order, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.orderRepo.List(userID, status, page, pageSize)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id int64, status string) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if err := s.orderRepo.UpdateStatus(id, status); err != nil {
		return nil, err
	}

	order.Status = status
	return order, nil
}
