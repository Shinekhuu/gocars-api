package redis

import (
	"context"

	"gocars-api/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var Client *redis.Client

func Init(cfg config.Config) {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.REDIS_ADDR,
		Password: cfg.REDIS_PASSWORD,
		DB:       cfg.REDIS_DB,
	})

	if _, err := Client.Ping(context.Background()).Result(); err != nil {
		zap.L().Warn("Redis connection failed — cache disabled", zap.Error(err))
		Client = nil
		return
	}

	zap.L().Info("Redis initialized successfully", zap.String("addr", cfg.REDIS_ADDR))
}