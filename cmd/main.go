package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: os.Getenv("DATABASE_URL"),
		},
	}

	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	logger.Info("connected to database", "dsn", os.Getenv("DATABASE_URL"))

	api := application{
		config: cfg,
		db:     conn,
	}

	if err := api.run(api.mount()); err != nil {
		slog.Error("server is failed to start", "error", err)
		os.Exit(1)
	}
}
