package support

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/testcontainers/testcontainers-go"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	urlSegmentsCount = 2
	redisImage       = "public.ecr.aws/docker/library/redis:7"
	attempts         = 10
)

type Redis struct {
	Client    *redis.Client
	Container *redisContainer.RedisContainer
	Host      string
	Port      string
}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) SetUp() {
	ctx := context.Background()

	var redisCtn *redisContainer.RedisContainer

	_, _, attemptErr := lo.AttemptWithDelay(attempts, 1*time.Second, func(_ int, _ time.Duration) error {
		ctn, err := redisContainer.RunContainer(ctx, testcontainers.WithImage(redisImage))
		if err != nil {
			return err
		}

		redisCtn = ctn

		return nil
	})
	if attemptErr != nil {
		panic(attemptErr)
	}

	host := lo.Must(redisCtn.Host(context.Background()))
	endpoint := lo.Must(redisCtn.Endpoint(context.Background(), ""))
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    endpoint,
	})

	parts := strings.Split(endpoint, ":")
	if len(parts) < urlSegmentsCount {
		panic(parts)
	}

	r.Host = host
	r.Port = parts[1]
	r.Container = redisCtn
	r.Client = client
}

func (r *Redis) TearDown() {
	lo.Must0(r.Container.Terminate(context.Background()))
}
