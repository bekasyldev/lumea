package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bekasyldev/lumea/internal/media/s3store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// os.Interrupt equal to syscal.SIGINT
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: os.Getenv("DATABASE_URL"),
		},
		s3: s3store.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    os.Getenv("S3_REGION"),
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		},
	}

	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	logger.Info("connected to database")

	store, err := s3store.New(ctx, cfg.s3)

	if err != nil {
		panic(err)
	}

	api := application{
		config:  cfg,
		db:      pool,
		storage: store,
	}

	if err := api.run(ctx, api.mount()); err != nil {
		slog.Error("server is failed to start", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
