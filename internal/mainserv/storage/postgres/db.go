package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drobyshevv/doc-service/internal/mainserv/config"
)

// NewPool создаёт и настраивает пул соединений PostgreSQL.
//
// Настройки пула:
//   - MaxConns: максимальное количество одновременных соединений
//   - MinConns: минимальное количество заранее открытых соединений
//   - MaxConnLifetime: максимальное время жизни соединения
//   - MaxConnIdleTime: максимальное время простоя соединения
//
// После создания выполняется Ping для проверки доступности БД.
func NewPool(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DBConnStr())
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
