package middlewares

import (
	"fmt"
	"main/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		lastError := c.Errors.Last()
		httpStatus, errorMessage, errCode, fieldsErr := mapErrorToResponse(lastError)

		errorResponse := gin.H{
			"code":       httpStatus,
			"status":     "error",
			"message":    errorMessage,
			"error_code": errCode,
		}
		if fieldsErr != nil {
			errorResponse["fields"] = fieldsErr
		}
		c.JSON(httpStatus, errorResponse)
		c.Abort()
	}
}
func mapErrorToResponse(err *gin.Error) (int, string, string, []string) {
	customErr, ok := err.Err.(*util.CustomError)
	if !ok {
		return http.StatusInternalServerError, err.Error(), "", nil
	}
	if ve, ok := customErr.Err.(validator.ValidationErrors); ok && customErr.ErrCode == util.ErrorValidationFailed {
		errorMessages := make([]string, 0)
		for _, fe := range ve {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' validation failed for tag '%s'", fe.Field(), fe.Tag()))
		}
		return customErr.Status, customErr.Message, customErr.ErrCode, errorMessages
	}

	return customErr.Status, customErr.Message, customErr.ErrCode, nil
}
