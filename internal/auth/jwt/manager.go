package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secretKey:       []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

type jwtCustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (m *Manager) GenerateAccessToken(userID uuid.UUID, role string) (string, error) {
	now := time.Now()

	claims := jwtCustomClaims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secretKey)
}

func (m *Manager) GenerateRefreshToken() (string, time.Time) {
	exp := time.Now().Add(m.refreshTokenTTL)

	return uuid.NewString(), exp
}

func (m *Manager) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwtCustomClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return m.secretKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwtCustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	return &AccessClaims{
		UserID: userID,
		Role:   claims.Role,
		Exp:    claims.ExpiresAt.Time.Unix(),
	}, nil
}
