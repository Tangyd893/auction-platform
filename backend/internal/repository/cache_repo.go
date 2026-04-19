package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *CacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *CacheRepository) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

func (r *CacheRepository) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.client.Subscribe(ctx, channel)
}

// 缓存 key 前缀
const (
	KeyPrefixItem     = "item:"
	KeyPrefixUser    = "user:"
	KeyPrefixBid     = "bid:"
	KeyPrefixOrder    = "order:"
	ChannelBidUpdate  = "bid_updates:"
)

func ItemKey(id int64) string        { return fmt.Sprintf("%s%d", KeyPrefixItem, id) }
func UserKey(id int64) string        { return fmt.Sprintf("%s%d", KeyPrefixUser, id) }
func BidKey(itemID int64) string     { return fmt.Sprintf("%s%d", KeyPrefixBid, itemID) }
func OrderKey(id int64) string       { return fmt.Sprintf("%s%d", KeyPrefixOrder, id) }
func BidUpdateChannel(itemID int64) string { return fmt.Sprintf("%s%d", ChannelBidUpdate, itemID) }
