package store

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("store: key not found")

type Redis struct{ Client *redis.Client }

func NewRedis() *Redis {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Redis{Client: rdb}
}

func (r *Redis) SetJTI(ctx context.Context, key, userID string, exp time.Time) error {
	return r.Client.Set(ctx, key, userID, time.Until(exp)).Err()
}

func (r *Redis) DelJTI(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// reads and deletes a JTI
func (r *Redis) ConsumeJTI(ctx context.Context, key string) (string, error) {
	userID, err := r.Client.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return userID, err
}
