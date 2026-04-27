package service

import (
	"context"
	"time"

	"auction-platform/internal/model"
)

// ============ Repository Interfaces（用于依赖注入和单元测试）============

type BidRepo interface {
	Create(bid *model.Bid) error
	GetByID(id int64) (*model.Bid, error)
	ListByItemID(itemID int64) ([]*model.Bid, error)
	ListByBuyerID(buyerID int64) ([]*model.Bid, error)
	CountByItemID(itemID int64) (int64, error)
	UpdateStatus(id int64, status string) error
	MarkItemBidsOutbid(itemID int64, exceptBidID int64) error
}

type ItemRepo interface {
	Create(item *model.Item) error
	GetByID(id int64) (*model.Item, error)
	List(status string, sellerID int64, keyword string, page, pageSize int) ([]*model.Item, int, error)
	Update(item *model.Item) error
	UpdatePrice(id, price int64) error
	UpdateStatus(id int64, status string) error
	Delete(id int64) error
}

type OrderRepo interface {
	Create(order *model.Order) error
	GetByID(id int64) (*model.Order, error)
	List(userID int64, status string, page, pageSize int) ([]*model.Order, int, error)
	UpdateStatus(id int64, status string) error
	GetByItemID(itemID int64) (*model.Order, error)
}

type UserRepo interface {
	Create(user *model.User) error
	GetByID(id int64) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	List(page, pageSize int) ([]*model.User, int, error)
	Update(user *model.User) error
	Delete(id int64) error
	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

type CacheRepo interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Publish(ctx context.Context, channel string, message interface{}) error
}
