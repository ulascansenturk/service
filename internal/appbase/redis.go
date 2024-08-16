package appbase

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	Client *redis.Client
}

func NewRedisService(host, port string) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	return &RedisService{Client: client}
}

func (s *RedisService) HealthCheck() error {
	return s.Client.Ping(context.Background()).Err()
}

func (s *RedisService) Shutdown() error {
	return s.Client.Close()
}
