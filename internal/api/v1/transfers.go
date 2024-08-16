package v1

import (
	"context"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/client"
	"net/http"
	"ulascansenturk/service/internal/api/server"
	"ulascansenturk/service/internal/constants"
	"ulascansenturk/service/internal/temporalworkflows"
	"ulascansenturk/service/internal/temporalworkflows/activities"
)

type TransfersService struct {
	transfersTaskQueueName string
	temporalClient         client.Client
}

func NewTransfersService(transfersTaskQueueName string, temporalClient client.Client) *TransfersService {
	return &TransfersService{transfersTaskQueueName: transfersTaskQueueName, temporalClient: temporalClient}
}

func (a *API) V1RunTransferWorkflow(w http.ResponseWriter, r *http.Request) {
	reqBody := new(server.V1RunTransferWorkflowJSONRequestBody)

	err := render.Bind(r, reqBody)
	if err != nil {
		server.BadRequestError(err, w, r)

		return
	}

	result, err := a.transfersService.RunRouteTransferWorkflow(r.Context(), reqBody)
	if err != nil {
		log.Err(err).Msg("transfer processing failed")

		server.ProcessingError(err, w, r)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, result)
}

func (s *TransfersService) RunRouteTransferWorkflow(
	ctx context.Context,
	reqBody *server.V1RunTransferWorkflowJSONRequestBody,
) (*activities.TransferResult, error) {
	params := temporalworkflows.TransferParams(reqBody.Data)
	workflowReferenceID := reqBody.Data.ReferenceId.String()

	ctx = context.WithValue(ctx, constants.ContextKeyWorkflowReferenceId.String(), workflowReferenceID)

	we, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        workflowReferenceID,
			TaskQueue: s.transfersTaskQueueName,
		},
		temporalworkflows.Transfer,
		&params,
	)
	if err != nil {
		return nil, err
	}

	var result activities.TransferResult

	err = we.Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
