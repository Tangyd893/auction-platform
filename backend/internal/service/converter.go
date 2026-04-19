package service

import (
	"context"

	"auction-platform/internal/model"
	pb "auction-platform/proto/gen/auction"
)

func toProtoUser(u *model.User) *pb.User {
	if u == nil {
		return nil
	}
	return &pb.User{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      pb.UserRole(pb.UserRole_value[u.Role]),
		CreatedAt: u.CreatedAt.Unix(),
	}
}

func toProtoUsers(users []*model.User) []*pb.User {
	result := make([]*pb.User, len(users))
	for i, u := range users {
		result[i] = toProtoUser(u)
	}
	return result
}

func toProtoItem(i *model.Item) *pb.Item {
	if i == nil {
		return nil
	}
	return &pb.Item{
		Id:           i.ID,
		Title:        i.Title,
		Description:  i.Description,
		ImageUrl:     i.ImageURL,
		StartPrice:   i.StartPrice,
		CurrentPrice: i.CurrentPrice,
		ReservePrice: i.ReservePrice,
		BidIncrement: i.BidIncrement,
		SellerId:     i.SellerID,
		Status:       pb.ItemStatus(pb.ItemStatus_value[i.Status]),
		StartTime:    i.StartTime.Unix(),
		EndTime:      i.EndTime.Unix(),
		CreatedAt:    i.CreatedAt.Unix(),
		UpdatedAt:    i.UpdatedAt.Unix(),
	}
}

func toProtoItems(items []*model.Item) []*pb.Item {
	result := make([]*pb.Item, len(items))
	for i, item := range items {
		result[i] = toProtoItem(item)
	}
	return result
}

func toProtoBid(b *model.Bid) *pb.Bid {
	if b == nil {
		return nil
	}
	return &pb.Bid{
		Id:        b.ID,
		ItemId:    b.ItemID,
		BuyerId:   b.BuyerID,
		Amount:    b.Amount,
		Status:    pb.BidStatus(pb.BidStatus_value[b.Status]),
		CreatedAt: b.CreatedAt.Unix(),
	}
}

func toProtoBids(bids []*model.Bid) []*pb.Bid {
	result := make([]*pb.Bid, len(bids))
	for i, b := range bids {
		result[i] = toProtoBid(b)
	}
	return result
}

func orderStatusToInt(s string) int32 {
	switch s {
	case "pending":
		return 0
	case "paid":
		return 1
	case "shipped":
		return 2
	case "completed":
		return 3
	case "cancelled":
		return 4
	default:
		return 0
	}
}

func toProtoOrder(o *model.Order) *pb.Order {
	if o == nil {
		return nil
	}
	var paidAt int64
	if o.PaidAt != nil {
		paidAt = o.PaidAt.Unix()
	}
	return &pb.Order{
		Id:         o.ID,
		ItemId:     o.ItemID,
		SellerId:   o.SellerID,
		BuyerId:    o.BuyerID,
		FinalPrice: o.FinalPrice,
		Status:     orderStatusToInt(o.Status),
		CreatedAt:  o.CreatedAt.Unix(),
		PaidAt:     paidAt,
	}
}

func toProtoOrders(orders []*model.Order) []*pb.Order {
	result := make([]*pb.Order, len(orders))
	for i, o := range orders {
		result[i] = toProtoOrder(o)
	}
	return result
}

// getUserIDFromContext 从 context 获取用户ID（需要 gateway 层注入）
func getUserIDFromContext(ctx context.Context) int64 {
	// TODO: 从 context 中提取 JWT claims 中的 user_id
	// 目前临时返回 1，便于测试
	return 1
}
