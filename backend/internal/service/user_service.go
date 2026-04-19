package service

import (
	"context"

	"auction-platform/internal/model"
)

type UserService struct {
	userRepo UserRepo
}

func NewUserService(userRepo UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) List(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.userRepo.List(page, pageSize)
}
