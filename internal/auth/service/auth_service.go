package service

import (
	"context"
	"errors"
	"time"

	"github.com/drobyshevv/doc-service/internal/auth/jwt"
	"github.com/drobyshevv/doc-service/internal/auth/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type AuthService struct {
	userRepo    UserRepository
	refreshRepo RefreshTokenRepository
	jwt         *jwt.Manager
}

func NewAuthService(
	userRepo UserRepository,
	refreshRepo RefreshTokenRepository,
	jwt *jwt.Manager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		jwt:         jwt,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, *model.TokenPair, error) {
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	access, err := s.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	refresh, exp := s.jwt.GenerateRefreshToken()

	tokenPair := &model.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp.Unix(),
	}

	_ = s.refreshRepo.Create(ctx, &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: exp,
	})

	return user, tokenPair, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, *model.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	access, err := s.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	refresh, exp := s.jwt.GenerateRefreshToken()

	tokenPair := &model.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp.Unix(),
	}

	_ = s.refreshRepo.Create(ctx, &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: exp,
	})

	return user, tokenPair, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.User, *model.TokenPair, error) {
	rt, err := s.refreshRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, err
	}
	if rt == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.refreshRepo.Delete(ctx, refreshToken)
		return nil, nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	_ = s.refreshRepo.Delete(ctx, refreshToken)

	access, err := s.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	refresh, exp := s.jwt.GenerateRefreshToken()

	newRT := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: exp,
	}

	if err := s.refreshRepo.Create(ctx, newRT); err != nil {
		return nil, nil, err
	}

	return user, &model.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp.Unix(),
	}, nil
}
