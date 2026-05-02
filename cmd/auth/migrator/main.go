package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/drobyshevv/doc-service/internal/mainserv/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.LoadConfig("configs/auth.yaml")
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	m, err := migrate.New(
		"file://migrations/auth",
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
