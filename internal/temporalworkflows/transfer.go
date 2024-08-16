package temporalworkflows

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
	"ulascansenturk/service/internal/api/server"
	"ulascansenturk/service/internal/temporalworkflows/activities"
)

type TransferParams server.TransferWorkflowParams

type TransferEnvConfig struct {
	TransferMutexTTLSeconds int `env:"TRANSFER_MUTEX_TTL_SECONDS" env-default:"300"`
}

func (p *TransferParams) sourceTransactionReferenceID() uuid.UUID {
	return getActivityReferenceID(p.ReferenceId, "transfer-source")
}

func (p *TransferParams) destinationTransactionReferenceID() uuid.UUID {
	return getActivityReferenceID(p.ReferenceId, "transfer-destination")
}

func (p *TransferParams) feeTransactionReferenceID() uuid.UUID {
	return getActivityReferenceID(p.ReferenceId, "transfer-fee")
}

func Transfer(ctx workflow.Context, params *TransferParams) (*activities.TransferResult, error) {
	var cfg TransferEnvConfig

	readCfgErr := cleanenv.ReadEnv(&cfg)
	if readCfgErr != nil {
		return nil, readCfgErr
	}

	releaseFunc, mutexErr := mutexLock(ctx, cfg, params)
	if mutexErr != nil {
		return nil, mutexErr
	}

	defer releaseFunc()

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	ctx = workflow.WithWorkflowID(ctx, getWorkflowReferenceID(params.ReferenceId).String())

	var (
		transactionOperations *activities.TransactionOperations
		transactionsResult    *activities.TransferResult
	)

	err := workflow.ExecuteActivity(ctx, transactionOperations.Transfer, activities.TransferParams{
		Amount:                            params.Amount,
		FeeAmount:                         params.FeeAmount,
		Metadata:                          params.Metadata,
		DestinationAccountID:              params.DestinationAccountID,
		SourceTransactionReferenceID:      params.sourceTransactionReferenceID(),
		DestinationTransactionReferenceID: params.destinationTransactionReferenceID(),
		FeeTransactionReferenceID:         params.feeTransactionReferenceID(),
		SourceAccountID:                   params.SourceAccountID,
	}).Get(ctx, &transactionsResult)
	if err != nil {
		return nil, err
	}

	return &activities.TransferResult{
		SourceTransactionReferenceID:      transactionsResult.SourceTransactionReferenceID,
		DestinationTransactionReferenceID: transactionsResult.DestinationTransactionReferenceID,
		FeeTransactionReferenceID:         transactionsResult.FeeTransactionReferenceID,
		FeeTransaction:                    transactionsResult.FeeTransaction,
		SourceTransaction:                 transactionsResult.SourceTransaction,
		DestinationTransaction:            transactionsResult.DestinationTransaction,
	}, nil

}

type MutexReleaseFunc func() error

func mutexLock(ctx workflow.Context, cfg TransferEnvConfig, params *TransferParams) (MutexReleaseFunc, error) {
	var (
		mutex *activities.Mutex
	)

	mutexCtx := workflow.WithActivityOptions(
		ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 2 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    5 * time.Second,
				BackoffCoefficient: 1.0,
				MaximumInterval:    10 * time.Second,
			},
		},
	)
	mutexParams := activities.MutexParams{
		Key:            fmt.Sprintf("transfers_mutex_%s", params.SourceAccountID.String()),
		OwnershipToken: params.ReferenceId.String(),
		TTL:            time.Duration(cfg.TransferMutexTTLSeconds) * time.Second,
	}

	acquireLockErr := workflow.ExecuteActivity(
		mutexCtx,
		mutex.AcquireLock,
		mutexParams,
	).Get(ctx, nil)
	if acquireLockErr != nil {
		return nil, acquireLockErr
	}

	return func() error {
		releaseLockErr := workflow.ExecuteActivity(
			mutexCtx,
			mutex.ReleaseLock,
			mutexParams,
		).Get(ctx, nil)

		return releaseLockErr
	}, nil
}

func getActivityReferenceID(workflowReference uuid.UUID, prefix string) uuid.UUID {
	return uuid.NewSHA1(
		uuid.NameSpaceDNS,
		[]byte(fmt.Sprintf("%s-%s", prefix, workflowReference.String())),
	)
}
