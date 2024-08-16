package server

import "net/http"

func (b *V1RunTransferWorkflowJSONRequestBody) Bind(_ *http.Request) error {
	return nil
}

func (b *V1CreateUserJSONRequestBody) Bind(_ *http.Request) error {
	return nil
}
