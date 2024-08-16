package server

import (
	"net/http"

	"github.com/go-chi/render"
)

const (
	badRequestErrorTitle = "BAD_REQUEST"
	processingErrorTitle = "PROCESSING_ERROR"
	timeoutErrorTitle    = "TIMEOUT"
	notFoundErrorTitle   = "NOT_FOUND"
)

func BadRequestError(badRequestErr error, w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusBadRequest

	errs := make([]Error, 0)

	err := Error{
		Code:   http.StatusText(statusCode),
		Detail: badRequestErr.Error(),
		Meta:   map[string]interface{}{},
		Status: statusCode,
		Title:  badRequestErrorTitle,
	}

	errs = append(errs, err)

	errResponse := ErrorResponse{Errors: errs}

	render.Status(r, statusCode)
	render.JSON(w, r, errResponse)
}

func ProcessingError(unprocessibleErr error, w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusUnprocessableEntity

	errs := make([]Error, 0)

	err := Error{
		Code:   http.StatusText(statusCode),
		Detail: unprocessibleErr.Error(),
		Meta:   map[string]interface{}{},
		Status: statusCode,
		Title:  processingErrorTitle,
	}

	errs = append(errs, err)

	errResponse := ErrorResponse{Errors: errs}

	render.Status(r, statusCode)
	render.JSON(w, r, errResponse)
}
