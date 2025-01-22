package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ErrorValidationFailed = "ERROR_01"
	ErrorNotFound         = "ERROR_02"
	ErrorInternal         = "ERROR_03" // Add internal error code
)

func HasContextError(ctx *gin.Context) bool {
	return len(ctx.Errors) > 0
}

type CustomError struct {
	Err     error
	Status  int
	ErrCode string
	Message string
}

func (e *CustomError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Err.Error()
}

func NewValidationError(err error, message string) *CustomError {
	return &CustomError{
		Err:     err,
		Status:  http.StatusBadRequest,
		Message: message,
		ErrCode: ErrorValidationFailed,
	}
}

func NewNotFoundError(err error, message string) *CustomError {
	return &CustomError{
		Err:     err,
		Status:  http.StatusNotFound,
		Message: message,
		ErrCode: ErrorNotFound,
	}
}

func NewInternalServerError(err error, message string) *CustomError {
	return &CustomError{
		Err:     err,
		Status:  http.StatusInternalServerError,
		Message: message,
		ErrCode: ErrorInternal,
	}
}
