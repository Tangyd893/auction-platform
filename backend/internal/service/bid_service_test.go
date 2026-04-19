package service

import (
	"context"
	"testing"
	"time"

	"auction-platform/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============ Mock Repositories ============

type MockBidRepository struct {
	mock.Mock
}

func (m *MockBidRepository) Create(bid *model.Bid) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockBidRepository) GetByID(id int64) (*model.Bid, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Bid), args.Error(1)
}

func (m *MockBidRepository) GetHighestBid(itemID int64) (*model.Bid, error) {
	args := m.Called(itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Bid), args.Error(1)
}

func (m *MockBidRepository) ListByItemID(itemID int64) ([]*model.Bid, error) {
	args := m.Called(itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Bid), args.Error(1)
}

func (m *MockBidRepository) ListByBuyerID(buyerID int64) ([]*model.Bid, error) {
	args := m.Called(buyerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Bid), args.Error(1)
}

func (m *MockBidRepository) CountByItemID(itemID int64) (int64, error) {
	args := m.Called(itemID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBidRepository) UpdateStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockBidRepository) MarkItemBidsOutbid(itemID int64, exceptBidID int64) error {
	args := m.Called(itemID, exceptBidID)
	return args.Error(0)
}

type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) Create(item *model.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockItemRepository) GetByID(id int64) (*model.Item, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Item), args.Error(1)
}

func (m *MockItemRepository) List(status string, sellerID int64, keyword string, page, pageSize int) ([]*model.Item, int, error) {
	args := m.Called(status, sellerID, keyword, page, pageSize)
	return args.Get(0).([]*model.Item), args.Int(1), args.Error(2)
}

func (m *MockItemRepository) Update(item *model.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockItemRepository) UpdatePrice(id, price int64) error {
	args := m.Called(id, price)
	return args.Error(0)
}

func (m *MockItemRepository) UpdateStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockItemRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) Publish(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}

// ============ Bid Service Tests ============

func TestPlaceBid_Success(t *testing.T) {
	mockBidRepo := new(MockBidRepository)
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewBidService(
		&BidRepositoryAdapter{mockBidRepo},
		&ItemRepositoryAdapter{mockItemRepo},
		&CacheRepositoryAdapter{mockCacheRepo},
	)

	now := time.Now()
	item := &model.Item{
		ID:           1,
		SellerID:     10,
		Status:       string(model.ItemStatusListed),
		StartPrice:   10000,
		CurrentPrice: 10000,
		BidIncrement: 100,
		StartTime:    now.Add(-1 * time.Hour),
		EndTime:      now.Add(1 * time.Hour),
	}

	prevBid := &model.Bid{
		ID:      2, // 不同的 ID
		ItemID:  1,
		BuyerID: 30,
		Amount:  10000,
		Status:  string(model.BidStatusActive),
	}

	mockItemRepo.On("GetByID", int64(1)).Return(item, nil)
	mockBidRepo.On("Create", mock.AnythingOfType("*model.Bid")).Return(nil).Run(func(args mock.Arguments) {
		b := args.Get(0).(*model.Bid)
		b.ID = 1
	})
	mockItemRepo.On("UpdatePrice", int64(1), int64(10100)).Return(nil)
	mockBidRepo.On("GetHighestBid", int64(1)).Return(prevBid, nil)
	mockBidRepo.On("UpdateStatus", int64(2), string(model.BidStatusOutbid)).Return(nil)

	ctx := context.Background()
	resultBid, currentPrice, isWinning, err := svc.PlaceBid(ctx, 1, 20, 10100)

	assert.NoError(t, err)
	assert.Equal(t, int64(10100), currentPrice)
	assert.True(t, isWinning)
	assert.NotNil(t, resultBid)

	mockItemRepo.AssertExpectations(t)
	mockBidRepo.AssertExpectations(t)
}

func TestPlaceBid_TooLow(t *testing.T) {
	mockBidRepo := new(MockBidRepository)
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewBidService(
		&BidRepositoryAdapter{mockBidRepo},
		&ItemRepositoryAdapter{mockItemRepo},
		&CacheRepositoryAdapter{mockCacheRepo},
	)

	now := time.Now()
	item := &model.Item{
		ID:           1,
		SellerID:     10,
		Status:       string(model.ItemStatusListed),
		CurrentPrice: 10000,
		BidIncrement: 100,
		StartTime:    now.Add(-1 * time.Hour),
		EndTime:      now.Add(1 * time.Hour),
	}

	mockItemRepo.On("GetByID", int64(1)).Return(item, nil)

	ctx := context.Background()
	_, _, _, err := svc.PlaceBid(ctx, 1, 20, 10050) // 低于最低加价 10100

	assert.Error(t, err)
	assert.Equal(t, ErrBidTooLow, err)
}

func TestPlaceBid_CannotBidOwnItem(t *testing.T) {
	mockBidRepo := new(MockBidRepository)
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewBidService(
		&BidRepositoryAdapter{mockBidRepo},
		&ItemRepositoryAdapter{mockItemRepo},
		&CacheRepositoryAdapter{mockCacheRepo},
	)

	now := time.Now()
	item := &model.Item{
		ID:        1,
		SellerID:  10,
		Status:    string(model.ItemStatusListed),
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(1 * time.Hour),
	}

	mockItemRepo.On("GetByID", int64(1)).Return(item, nil)

	ctx := context.Background()
	_, _, _, err := svc.PlaceBid(ctx, 1, 10, 10100) // buyerID == sellerID

	assert.Error(t, err)
	assert.Equal(t, ErrCannotBidOwn, err)
}

func TestPlaceBid_AuctionEnded(t *testing.T) {
	mockBidRepo := new(MockBidRepository)
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewBidService(
		&BidRepositoryAdapter{mockBidRepo},
		&ItemRepositoryAdapter{mockItemRepo},
		&CacheRepositoryAdapter{mockCacheRepo},
	)

	now := time.Now()
	item := &model.Item{
		ID:        1,
		SellerID:  10,
		Status:    string(model.ItemStatusListed),
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now.Add(-1 * time.Hour), // 已结束
	}

	mockItemRepo.On("GetByID", int64(1)).Return(item, nil)

	ctx := context.Background()
	_, _, _, err := svc.PlaceBid(ctx, 1, 20, 10100)

	assert.Error(t, err)
	assert.Equal(t, ErrAuctionEnded, err)
}

func TestPlaceBid_ItemNotActive(t *testing.T) {
	mockBidRepo := new(MockBidRepository)
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewBidService(
		&BidRepositoryAdapter{mockBidRepo},
		&ItemRepositoryAdapter{mockItemRepo},
		&CacheRepositoryAdapter{mockCacheRepo},
	)

	now := time.Now()
	item := &model.Item{
		ID:        1,
		SellerID:  10,
		Status:    string(model.ItemStatusSold), // 已售出
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(1 * time.Hour),
	}

	mockItemRepo.On("GetByID", int64(1)).Return(item, nil)

	ctx := context.Background()
	_, _, _, err := svc.PlaceBid(ctx, 1, 20, 10100)

	assert.Error(t, err)
	assert.Equal(t, ErrItemNotActive, err)
}

// ============ Item Service Tests ============

func TestCreateItem_EndTimeInPast(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewItemService(&ItemRepositoryAdapter{mockItemRepo}, &CacheRepositoryAdapter{mockCacheRepo})

	req := &CreateItemRequest{
		Title:      "Test",
		StartPrice: 10000,
		StartTime:  time.Now().Add(-1*time.Hour).Unix(),
		EndTime:    time.Now().Add(-30*time.Minute).Unix(), // 结束时间在过去
	}

	ctx := context.Background()
	_, err := svc.Create(ctx, req, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end time")
}

func TestCreateItem_StartAfterEnd(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewItemService(&ItemRepositoryAdapter{mockItemRepo}, &CacheRepositoryAdapter{mockCacheRepo})

	req := &CreateItemRequest{
		Title:      "Test",
		StartPrice: 10000,
		StartTime:  time.Now().Add(2*time.Hour).Unix(),
		EndTime:    time.Now().Add(1 * time.Hour).Unix(), // 开始在结束之后
	}

	ctx := context.Background()
	_, err := svc.Create(ctx, req, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start time")
}

func TestCancelItem_AlreadySold(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockCacheRepo := new(MockCacheRepository)

	svc := NewItemService(&ItemRepositoryAdapter{mockItemRepo}, &CacheRepositoryAdapter{mockCacheRepo})

	item := &model.Item{
		ID:       1,
		SellerID: 1,
		Status:   string(model.ItemStatusSold),
	}

	mockItemRepo.On("GetByID", int64(1)).Return(item, nil)

	ctx := context.Background()
	_, err := svc.Cancel(ctx, 1, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel")
}

// ============ Adapter（让 mock 适配 service 层接口）============

type BidRepositoryAdapter struct {
	mock *MockBidRepository
}

func (a *BidRepositoryAdapter) Create(bid *model.Bid) error {
	return a.mock.Create(bid)
}
func (a *BidRepositoryAdapter) GetByID(id int64) (*model.Bid, error) {
	return a.mock.GetByID(id)
}
func (a *BidRepositoryAdapter) GetHighestBid(itemID int64) (*model.Bid, error) {
	return a.mock.GetHighestBid(itemID)
}
func (a *BidRepositoryAdapter) ListByItemID(itemID int64) ([]*model.Bid, error) {
	return a.mock.ListByItemID(itemID)
}
func (a *BidRepositoryAdapter) ListByBuyerID(buyerID int64) ([]*model.Bid, error) {
	return a.mock.ListByBuyerID(buyerID)
}
func (a *BidRepositoryAdapter) CountByItemID(itemID int64) (int64, error) {
	return a.mock.CountByItemID(itemID)
}
func (a *BidRepositoryAdapter) UpdateStatus(id int64, status string) error {
	return a.mock.UpdateStatus(id, status)
}
func (a *BidRepositoryAdapter) MarkItemBidsOutbid(itemID int64, exceptBidID int64) error {
	return a.mock.MarkItemBidsOutbid(itemID, exceptBidID)
}

type ItemRepositoryAdapter struct {
	mock *MockItemRepository
}

func (a *ItemRepositoryAdapter) Create(item *model.Item) error {
	return a.mock.Create(item)
}
func (a *ItemRepositoryAdapter) GetByID(id int64) (*model.Item, error) {
	return a.mock.GetByID(id)
}
func (a *ItemRepositoryAdapter) List(status string, sellerID int64, keyword string, page, pageSize int) ([]*model.Item, int, error) {
	return a.mock.List(status, sellerID, keyword, page, pageSize)
}
func (a *ItemRepositoryAdapter) Update(item *model.Item) error {
	return a.mock.Update(item)
}
func (a *ItemRepositoryAdapter) UpdatePrice(id, price int64) error {
	return a.mock.UpdatePrice(id, price)
}
func (a *ItemRepositoryAdapter) UpdateStatus(id int64, status string) error {
	return a.mock.UpdateStatus(id, status)
}
func (a *ItemRepositoryAdapter) Delete(id int64) error {
	return a.mock.Delete(id)
}

type CacheRepositoryAdapter struct {
	mock *MockCacheRepository
}

func (a *CacheRepositoryAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return a.mock.Set(ctx, key, value, expiration)
}
func (a *CacheRepositoryAdapter) Get(ctx context.Context, key string, dest interface{}) error {
	return a.mock.Get(ctx, key, dest)
}
func (a *CacheRepositoryAdapter) Delete(ctx context.Context, key string) error {
	return a.mock.Delete(ctx, key)
}
func (a *CacheRepositoryAdapter) Publish(ctx context.Context, channel string, message interface{}) error {
	return a.mock.Publish(ctx, channel, message)
}
