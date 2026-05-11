package jwt

import "github.com/google/uuid"

type AccessClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Exp    int64     `json:"exp"`
}
