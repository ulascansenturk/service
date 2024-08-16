package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"ulascansenturk/service/internal/api/server"
	v1 "ulascansenturk/service/internal/api/v1"
)

// InitRoutes initializes the routes for the API.
func InitRoutes(router *chi.Mux, si *Routes) {
	server.HandlerFromMux(si, router)
}

// NewRoutes creates a new instance of Routes.
func NewRoutes(apiV1 *v1.API) *Routes {
	return &Routes{
		v1: apiV1,
	}
}

// Routes is the wrapper for all the versions of the API defined by server.ServerInterface.
type Routes struct {
	v1 *v1.API
}

func (a *Routes) V1CreateUser(w http.ResponseWriter, r *http.Request) {
	a.v1.V1CreateUser(w, r)
}

func (a *Routes) V1RunTransferWorkflow(w http.ResponseWriter, r *http.Request) {
	a.v1.V1RunTransferWorkflow(w, r)
}
