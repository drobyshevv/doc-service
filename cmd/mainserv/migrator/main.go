// Command migrate управляет миграциями базы данных.
//
// Поддерживаемые команды:
//   - up
//   - down
//   - step N
//   - force N
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/drobyshevv/doc-service/internal/mainserv/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxRetries    = 10
	retryInterval = 2 * time.Second
)

func waitForDB(connStr string) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		pool, err := pgxpool.New(ctx, connStr)
		if err == nil {
			err = pool.Ping(ctx)
			pool.Close()
		}

		if err == nil {
			fmt.Println("postgres is ready")
			return nil
		}

		lastErr = err
		fmt.Printf("waiting for postgres... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("postgres not ready after retries: %w", lastErr)
}

func main() {
	cfg, err := config.LoadConfig("configs/mainserv.yaml")
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	if err := waitForDB(cfg.DBConnStr()); err != nil {
		panic(err)
	}

	m, err := migrate.New(
		"file://migrations/mainserv",
		cfg.DBConnStr(),
	)
	if err != nil {
		panic(fmt.Errorf("create migrate instance: %w", err))
	}
	defer m.Close()

	if len(os.Args) < 2 {
		panic("command required: up | down | step N | force N")
	}

	command := os.Args[1]

	switch command {

	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}
		fmt.Println("migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}
		fmt.Println("all migrations rolled back")

	case "step":
		if len(os.Args) < 3 {
			panic("step requires number argument")
		}

		steps, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}

		if err := m.Steps(steps); err != nil && err != migrate.ErrNoChange {
			panic(err)
		}

		fmt.Printf("migrated %d steps\n", steps)

	case "force":
		if len(os.Args) < 3 {
			panic("force requires version argument")
		}

		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}

		if err := m.Force(version); err != nil {
			panic(err)
		}

		fmt.Printf("migration version forced to %d\n", version)

	default:
		panic("unknown command: use up | down | step N | force N")
	}
}
