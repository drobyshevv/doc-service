package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/drobyshevv/doc-service/internal/auth/model"
	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, t *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		t.ID,
		t.UserID,
		t.Token,
		t.ExpiresAt,
	)

	return err
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var t model.RefreshToken

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&t.ID,
		&t.UserID,
		&t.Token,
		&t.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

func (r *RefreshTokenRepository) Delete(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// TODO: для logout
func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
