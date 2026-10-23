package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/bekasyldev/lumea/internal/media/s3store"
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
		s3: s3store.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    os.Getenv("S3_REGION"),
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		},
	}

	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	logger.Info("connected to database", "dsn", os.Getenv("DATABASE_URL"))

	store, err := s3store.New(ctx, cfg.s3)

	if err != nil {
		panic(err)
	}

	api := application{
		config:  cfg,
		db:      conn,
		storage: store,
	}

	if err := api.run(api.mount()); err != nil {
		slog.Error("server is failed to start", "error", err)
		os.Exit(1)
	}
}
