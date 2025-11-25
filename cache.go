package main

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type Cache struct {
	Client *redis.Client
	Ttl    time.Duration
}

func NewCache(addr string) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		err := rdb.Ping(ctx).Err()
		if err == nil {
			break
		}
		log.Warn("waiting for cache server...")
		time.Sleep(500 * time.Millisecond)
	}

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return &Cache{
		Client: rdb,
		Ttl:    30 * time.Second,
	}, nil
}

const OTP_KEY = "otp:"

func (cache *Cache) SetOTP(username string, otp string) error {
	ctx := context.Background()
	key := OTP_KEY + username
	return cache.Client.Set(ctx, key, otp, cache.Ttl).Err()
}

func (cache *Cache) GetOTP(username string) (string, error) {
	ctx := context.Background()
	key := OTP_KEY + username
	return cache.Client.Get(ctx, key).Result()
}

func (cache *Cache) Close() error {
	return cache.Client.Close()
}
