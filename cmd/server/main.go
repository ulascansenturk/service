package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"go.temporal.io/sdk/worker"
	"net/http"
	"ulascansenturk/service/cmd/utils"

	"ulascansenturk/service/internal/appbase"
)

const (
	serviceName = "serviceName"
)

func main() {
	ctx, mainCtxStop := context.WithCancel(context.Background())

	app := appbase.New(
		appbase.Init(serviceName),
		appbase.WithDependencyInjector(),
	)
	defer app.Shutdown()

	wrk := do.MustInvokeNamed[worker.Worker](app.Injector, "transactions")

	router := buildRouter(app)

	httpServer := &http.Server{
		Addr:              app.Config.ServerAddress,
		Handler:           router,
		ReadHeaderTimeout: app.Config.HTTPTimeoutDuration(),
	}

	utils.HandleSignals(ctx, mainCtxStop, func() {
		shutdownErr := httpServer.Shutdown(ctx)
		if shutdownErr != nil {
			log.Fatal().Err(shutdownErr).Msg("server shutdown failed")
		}
	})

	log.Info().Msgf("started server on %s", app.Config.ServerAddress)

	go func() {
		err := wrk.Run(worker.InterruptCh())
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("temporal worker started !")
	}()

	serverErr := httpServer.ListenAndServe()
	if serverErr != nil {
		log.Err(serverErr).Msg("server stopped")
	}
	<-ctx.Done()
}
