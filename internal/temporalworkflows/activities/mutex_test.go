//go:build tests_unit

package activities

import (
	"context"
	"testing"
	"time"
	"ulascansenturk/service/internal/api/testutils/support"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/ulascansenturk/service/internal/api/testutils/support"
)

const maxContainerCreationAttempts = 10

type testSuiteMutex struct {
	suite.Suite

	mutex          *Mutex
	redisContainer *redisContainer.RedisContainer
	redisClient    *redis.Client
	rd             *support.Redis
}

func (s *testSuiteMutex) SetupTest() {
	s.createRedisContainer()

	endpoint := lo.Must(s.redisContainer.Endpoint(context.Background(), ""))

	s.redisClient = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    endpoint,
	})

	pool := goredis.NewPool(s.redisClient)
	locker := redsync.New(pool)

	s.mutex = NewMutex(locker)
}

func (s *testSuiteMutex) TearDownTest() {
	s.rd.TearDown()
}

func (s *testSuiteMutex) createRedisContainer() {
	s.rd = support.NewRedis()
	s.rd.SetUp()

	s.redisContainer = s.rd.Container
}

func TestMutex(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(testSuiteMutex))
}

func (s *testSuiteMutex) TestMutex_AcquireLock() {
	ctx := context.Background()

	s.Run("when the lock is acquired successfully", func() {
		params := MutexParams{
			Key:            uuid.New().String(),
			OwnershipToken: "owner-1",
			TTL:            1 * time.Minute,
		}

		err := s.mutex.AcquireLock(context.Background(), params)

		s.NoError(err)

		val, getKeyErr := s.redisClient.Get(ctx, params.Key).Result()

		s.Require().NoError(getKeyErr)

		s.Equal(params.OwnershipToken, val)

	})

	s.Run("when the lock cannot be acquired", func() {
		params := MutexParams{
			Key:            uuid.New().String(),
			OwnershipToken: "owner-1",
			TTL:            1 * time.Minute,
		}

		err := s.mutex.AcquireLock(context.Background(), params)

		s.NoError(err)

		val, getKeyErr := s.redisClient.Get(ctx, params.Key).Result()

		s.Require().NoError(getKeyErr)

		s.Equal(params.OwnershipToken, val)

		secondErr := s.mutex.AcquireLock(context.Background(), params)

		s.EqualError(secondErr, "redsync: failed to acquire lock")
	})
}

func (s *testSuiteMutex) TestMutex_ReleaseLock() {
	ctx := context.Background()

	s.Run("when lock is released successfully", func() {
		params := MutexParams{
			Key:            uuid.New().String(),
			OwnershipToken: "owner-1",
			TTL:            1 * time.Minute,
		}

		acquireLockErr := s.mutex.AcquireLock(ctx, params)
		s.Require().NoError(acquireLockErr)

		err := s.mutex.ReleaseLock(ctx, params)

		s.NoError(err)

		val, getKeyErr := s.redisClient.Get(ctx, params.Key).Result()

		s.Equal("", val)
		s.EqualError(getKeyErr, redis.Nil.Error())
	})

	s.Run("when releasing returns an error during the initial lock", func() {
		shortCtx, cancelShortCtx := context.WithCancel(ctx)
		params := MutexParams{
			Key:            uuid.New().String(),
			OwnershipToken: "owner-1",
			TTL:            1 * time.Minute,
		}

		acquireLockErr := s.mutex.AcquireLock(ctx, params)
		s.Require().NoError(acquireLockErr)

		cancelShortCtx()

		err := s.mutex.ReleaseLock(shortCtx, params)

		s.ErrorIs(err, redsync.ErrFailed)

		val, getKeyErr := s.redisClient.Get(ctx, params.Key).Result()

		s.Equal("owner-1", val)
		s.NoError(getKeyErr)
	})
}
