package cache

import (
	"KnowLedger/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisUniversalClient(cfg *config.Config) redis.UniversalClient {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{cfg.Storage.Redis.Addr},
		Password: cfg.Storage.Redis.Password,
		DB:       cfg.Storage.Redis.DB,
	})
	return client
}

func NewRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Storage.Redis.Addr,
		Password: cfg.Storage.Redis.Password,
		DB:       cfg.Storage.Redis.DB,
	})
	return client
}
