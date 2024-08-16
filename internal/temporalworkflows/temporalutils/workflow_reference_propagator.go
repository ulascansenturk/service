package temporalutils

import (
	"context"
	"github.com/rs/zerolog"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"os"
	"ulascansenturk/service/internal/constants"
)

type WorkflowContextPropagator struct {
	logger *zerolog.Logger
}

func NewWorkflowContextPropagator(logger *zerolog.Logger) *WorkflowContextPropagator {
	return &WorkflowContextPropagator{logger: logger}
}

func GetWorkflowCtxLogger(ctx context.Context) zerolog.Logger {
	logger, ok := ctx.Value(constants.ContextKeyWorkflowLogger).(zerolog.Logger)
	if !ok {
		return zerolog.New(os.Stdout)
	}

	return logger
}

func ContextWithLogger(ctx context.Context) context.Context {
	logger := GetWorkflowCtxLogger(ctx)

	return logger.WithContext(ctx)
}

func (p *WorkflowContextPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	wfReferenceIDErr := p.setCtxPayloadByKey(ctx, writer, constants.ContextKeyWorkflowReferenceId)
	if wfReferenceIDErr != nil {
		return wfReferenceIDErr
	}

	loggerErr := p.setCtxPayloadByKey(ctx, writer, constants.ContextKeyWorkflowLogger)
	if loggerErr != nil {
		return loggerErr
	}

	return nil
}

func (p *WorkflowContextPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	wfReferenceIDErr := p.setWorkflowCtxPayloadByKey(ctx, writer, constants.ContextKeyWorkflowReferenceId)
	if wfReferenceIDErr != nil {
		return wfReferenceIDErr
	}

	loggerErr := p.setWorkflowCtxPayloadByKey(ctx, writer, constants.ContextKeyWorkflowLogger)
	if loggerErr != nil {
		return loggerErr
	}

	return nil
}

func (p *WorkflowContextPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	logger := *p.logger

	if value, ok := reader.Get(constants.ContextKeyWorkflowReferenceId.String()); ok {
		var workflowContext constants.WorkflowContext
		if err := converter.GetDefaultDataConverter().FromPayload(value, &workflowContext); err != nil {
			return ctx, nil
		}

		ctx = context.WithValue(ctx, constants.ContextKeyWorkflowReferenceId, workflowContext)
		ctx = context.WithValue(ctx, constants.ContextKeyWorkflowLogger, logger)
	}

	ctx = logger.WithContext(ctx)

	return ctx, nil
}

func (p *WorkflowContextPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(constants.ContextKeyWorkflowReferenceId.String()); ok {
		var workflowContext constants.WorkflowContext
		if err := converter.GetDefaultDataConverter().FromPayload(value, &workflowContext); err != nil {
			return ctx, err
		}

		ctx = workflow.WithValue(ctx, constants.ContextKeyWorkflowReferenceId, workflowContext)
	}

	return ctx, nil
}

func (p *WorkflowContextPropagator) setCtxPayloadByKey(ctx context.Context, writer workflow.HeaderWriter, key constants.ContextKey) error {
	value := ctx.Value(key)

	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}

	writer.Set(key.String(), payload)

	return nil
}

func (p *WorkflowContextPropagator) setWorkflowCtxPayloadByKey(ctx workflow.Context, writer workflow.HeaderWriter, key constants.ContextKey) error {
	value := ctx.Value(key)

	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}

	writer.Set(key.String(), payload)

	return nil
}
