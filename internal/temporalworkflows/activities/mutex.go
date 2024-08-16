package activities

import (
	"context"
	"errors"

	"github.com/go-redsync/redsync/v4"
	"github.com/gookit/goutil/errorx"
	"github.com/rs/zerolog/log"
	"time"
)

type Mutex struct {
	locker *redsync.Redsync
}

func NewMutex(locker *redsync.Redsync) *Mutex {
	return &Mutex{locker: locker}
}

type MutexParams struct {
	// Key — Redis key used for the lock
	Key string
	// OwnershipToken — Redis value used for transferring lock ownership.
	// This is needed to release the lock from a difference context or process.
	OwnershipToken string
	// TTL — Definitive expiration period for the lock after which the lock is release automatically
	TTL time.Duration
}

// AcquireLock tries to acquire the lock with provided parameters once.
// Returns an error if the lock is not available.
func (m *Mutex) AcquireLock(ctx context.Context, params MutexParams) error {
	mutex := m.locker.NewMutex(
		params.Key,
		redsync.WithExpiry(params.TTL),
		redsync.WithGenValueFunc(func() (string, error) {
			return params.OwnershipToken, nil
		}),
	)

	lockErr := mutex.TryLockContext(ctx)
	if lockErr != nil {
		log.Ctx(ctx).Err(lockErr).Msg("Mutex#AcquireLock: TryLock error")

		return lockErr
	}

	return nil
}

// ReleaseLock acquires the lock first and then unlocks it immediately.
func (m *Mutex) ReleaseLock(ctx context.Context, params MutexParams) error {
	mutex := m.locker.NewMutex(
		params.Key,
		redsync.WithGenValueFunc(func() (string, error) {
			return params.OwnershipToken, nil
		}),
	)

	// Need to acquire lock first otherwise unlocking won't work
	lockErr := mutex.LockContext(ctx)
	if lockErr != nil {
		log.Ctx(ctx).Err(lockErr).Msg("Mutex#ReleaseLock: Lock error")

		return errorx.Wrap(lockErr, "TryLockContext")
	}

	ok, unlockErr := mutex.UnlockContext(ctx)
	if unlockErr != nil {
		return unlockErr
	}

	if !ok {
		return errors.New("release lock failed")
	}

	return nil
}
