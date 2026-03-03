package rdb

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr        string
	Password    string
	User        string
	DB          int
	MaxRetries  int
	DialTimeout time.Duration
	Timeout     time.Duration
}

func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		Username:     cfg.User,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
