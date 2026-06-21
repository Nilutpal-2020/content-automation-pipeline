package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(addr, password string) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}
