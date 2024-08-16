package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/samber/do"

	"ulascansenturk/service/internal/api"
	"ulascansenturk/service/internal/appbase"
)

func buildRouter(app *appbase.AppBase) *chi.Mux {
	mux := do.MustInvokeNamed[*chi.Mux](app.Injector, appbase.InjectorApplicationRouter)
	routes := do.MustInvoke[*api.Routes](app.Injector)

	api.InitRoutes(mux, routes)

	return mux
}
