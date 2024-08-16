package appbase

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"go.temporal.io/sdk/worker"
	"net/http"
	"os"
	"time"
	"ulascansenturk/service/internal/accounts"
	"ulascansenturk/service/internal/api"
	v1 "ulascansenturk/service/internal/api/v1"
	"ulascansenturk/service/internal/temporalworkflows"
	"ulascansenturk/service/internal/temporalworkflows/activities"
	"ulascansenturk/service/internal/temporalworkflows/temporalutils"
	"ulascansenturk/service/internal/transactions"
	"ulascansenturk/service/internal/users"
	"ulascansenturk/service/openapi"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/samber/do"
	"github.com/samber/lo"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"
	zerologAdapter "logur.dev/adapter/zerolog"
	"logur.dev/logur"

	"ulascansenturk/service/internal/api/server"
)

const (
	ldWaitDuration     = 60 * time.Second
	testLdWaitDuration = 0
)

func NewInjector(serviceName string, cfg *Config) *do.Injector {
	injector := do.New()

	do.Provide(injector, func(i *do.Injector) (*zerolog.Logger, error) {
		logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
		if err != nil {
			return nil, err
		}

		logger := zerolog.New(os.Stdout).
			Level(logLevel).
			With().
			Str("service_name", serviceName).
			Timestamp().
			Logger()

		return &logger, nil
	})
	do.Provide(injector, func(i *do.Injector) (*TemporalService, error) {

		temporalLogger := do.MustInvoke[*zerolog.Logger](i)
		adaptedLogger := logur.LoggerToKV(zerologAdapter.New(*temporalLogger))

		logger := do.MustInvoke[*zerolog.Logger](i)
		loggerCtxPropagator := temporalutils.NewWorkflowContextPropagator(logger)

		return NewTemporalService(client.Options{
			Logger:    adaptedLogger,
			HostPort:  cfg.TemporalHost,
			Namespace: cfg.TemporalNamespace,
			ContextPropagators: []workflow.ContextPropagator{
				loggerCtxPropagator,
			},
		})
	})
	do.ProvideNamed(injector, InjectorOpenAPIValidationMiddleware, func(i *do.Injector) (*openapi.ValidationMiddleware, error) {
		switch cfg.Env {
		case "test":
			return openapi.NewValidationMiddleware(
				openapi.WithDoc(lo.Must(server.GetSwagger())),
			), nil
		default:
			return openapi.NewValidationMiddleware(
				openapi.WithDoc(lo.Must(server.GetSwagger())),
				openapi.WithKinOpenAPIDefaults(),
			), nil
		}
	})
	do.ProvideNamed(injector, InjectorApplicationRouter, func(i *do.Injector) (*chi.Mux, error) {
		logger := do.MustInvoke[*zerolog.Logger](i)
		openAPIValidation := do.MustInvokeNamed[*openapi.ValidationMiddleware](i, InjectorOpenAPIValidationMiddleware)

		gormDB := do.MustInvokeNamed[*gorm.DB](injector, InjectorDatabase)

		return NewRouterMux(serviceName, logger, openAPIValidation, cfg.HTTPTimeoutDuration(), gormDB), nil
	})
	do.ProvideNamed(injector, InjectorDatabase, func(i *do.Injector) (*gorm.DB, error) {
		credentials := Credentials{
			Name:            cfg.ApplicationDatabaseName,
			Password:        cfg.ApplicationDatabasePassword,
			PrimaryHost:     cfg.ApplicationPrimaryDatabaseHost,
			ReadReplicaHost: cfg.ApplicationReadReplicaDatabaseHost,
			User:            cfg.ApplicationDatabaseUser,
			Port:            cfg.ApplicationDatabasePort,
		}
		connCfg := ConnectionConfig{
			MaxLifetime:        5 * time.Minute,
			MaxIdleTime:        3 * time.Minute,
			MaxOpenConnections: 25,
			MaxIdleConnections: 25,
		}

		database, err := NewPostgres(
			WithCredentials(&credentials),
			WithConnectionConfig(&connCfg),
			WithServiceName(fmt.Sprintf("%s.database.application", serviceName)),
		)
		if err != nil {
			return nil, err
		}

		return database.GormDB, nil
	})

	do.Provide(injector, func(i *do.Injector) (*validator.Validate, error) {
		return validator.New(), nil
	})
	//Repos

	do.Provide(injector, func(i *do.Injector) (*accounts.SQLRepository, error) {
		gormDB := do.MustInvokeNamed[*gorm.DB](injector, InjectorDatabase)
		return accounts.NewSQLRepository(gormDB), nil
	})

	do.Provide(injector, func(i *do.Injector) (*transactions.SQLRepository, error) {
		gormDB := do.MustInvokeNamed[*gorm.DB](injector, InjectorDatabase)
		return transactions.NewSQLRepository(gormDB), nil
	})

	do.Provide(injector, func(i *do.Injector) (*users.SQLRepository, error) {
		gormDB := do.MustInvokeNamed[*gorm.DB](injector, InjectorDatabase)
		return users.NewSQLRepository(gormDB), nil
	})

	//Services

	do.Provide(injector, func(i *do.Injector) (*users.UserServiceImpl, error) {
		userRepo := do.MustInvoke[*users.SQLRepository](i)

		return users.NewUserService(userRepo), nil
	})

	do.Provide(injector, func(i *do.Injector) (*transactions.TransactionServiceImpl, error) {
		transactionsRepo := do.MustInvoke[*transactions.SQLRepository](i)

		validation := do.MustInvoke[*validator.Validate](i)

		return transactions.NewTransactionService(transactionsRepo, validation), nil
	})

	do.Provide(injector, func(i *do.Injector) (*accounts.AccountServiceImpl, error) {
		accountRepo := do.MustInvoke[*accounts.SQLRepository](i)

		validation := do.MustInvoke[*validator.Validate](i)

		return accounts.NewUserBankAccountService(accountRepo, validation), nil
	})

	do.Provide(injector, func(i *do.Injector) (*transactions.FinderOrCreatorService, error) {
		transactionsRepo := do.MustInvoke[*transactions.SQLRepository](i)

		accountsRepo := do.MustInvoke[*accounts.SQLRepository](i)

		validation := do.MustInvoke[*validator.Validate](i)

		return transactions.NewFinderOrCreatorService(transactionsRepo, accountsRepo, validation), nil

	})

	do.Provide(injector, func(i *do.Injector) (*v1.API, error) {

		temporalService := do.MustInvoke[*TemporalService](i)

		userServ := do.MustInvoke[*users.UserServiceImpl](i)

		accountsServ := do.MustInvoke[*accounts.AccountServiceImpl](i)

		transferService := v1.NewTransfersService(cfg.TemporalTransfersTaskQueueName, temporalService.Client)

		userService := v1.NewUsersService(userServ, accountsServ)
		return v1.NewAPI(transferService, userService), nil
	})

	do.Provide(injector, func(i *do.Injector) (*api.Routes, error) {
		v1API := do.MustInvoke[*v1.API](i)

		return api.NewRoutes(v1API), nil
	})

	do.Provide(injector, func(i *do.Injector) (*redsync.Redsync, error) {
		redisService := do.MustInvoke[*RedisService](i)
		pool := goredis.NewPool(redisService.Client)

		return redsync.New(pool), nil
	})

	do.Provide(injector, func(i *do.Injector) (*RedisService, error) {
		return NewRedisService(cfg.RedisEndpoint, cfg.RedisPort), nil
	})

	// API clients
	do.ProvideNamed(injector, InjectorDefaultHTTPClient, func(i *do.Injector) (http.Client, error) {
		return http.Client{}, nil
	})

	do.Provide(injector, func(i *do.Injector) (*activities.Mutex, error) {
		locker := do.MustInvoke[*redsync.Redsync](i)

		return activities.NewMutex(locker), nil
	})

	do.Provide(injector, func(i *do.Injector) (*activities.TransactionOperations, error) {
		finderOrCreatorService := do.MustInvoke[*transactions.FinderOrCreatorService](i)

		transactionsService := do.MustInvoke[*transactions.TransactionServiceImpl](i)

		accountsService := do.MustInvoke[*accounts.AccountServiceImpl](i)

		return activities.NewTransactionOperations(finderOrCreatorService, transactionsService, accountsService), nil
	})

	do.ProvideNamed(injector, "transactions", func(i *do.Injector) (worker.Worker, error) {
		wrk := worker.New(
			do.MustInvoke[*TemporalService](i).Client,
			cfg.TemporalTransfersTaskQueueName,
			worker.Options{},
		)

		transactionActivities := do.MustInvoke[*activities.TransactionOperations](i)

		mutexActivity := do.MustInvoke[*activities.Mutex](i)

		wrk.RegisterActivity(transactionActivities)
		wrk.RegisterActivity(mutexActivity)
		wrk.RegisterWorkflow(temporalworkflows.Transfer)

		return wrk, nil
	})

	return injector
}
