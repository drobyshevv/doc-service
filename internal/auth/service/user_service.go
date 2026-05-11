package service

import (
	"context"

	"github.com/drobyshevv/doc-service/internal/auth/model"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}
