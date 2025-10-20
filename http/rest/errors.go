package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type APIError interface {
	GetData() (int, string, string, []FieldError)
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field       string `json:"field"`
	Description string `json:"description,omitempty"`
}

// ErrorResponse is the form used for API responses from failures in the API.
// This is what is exposed to the client
type ErrorResponse struct {
	Error *ErrorPayload `json:"error"`
}

// ErrorPayload always attached ErrorResponse
type ErrorPayload struct {
	Code    string       `json:"code,omitempty"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// Error is used to pass an error during the request through the
// application with web specific context and internal error `Err`
// that could be logged.
type Error struct {
	Err         error
	Status      int
	Code        string
	Description string
	Fields      []FieldError
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected web.
func NewRequestError(err error, status int, code string, publicMsg string) *Error {
	return &Error{
		Err:         err,
		Status:      status,
		Code:        code,
		Description: publicMsg,
	}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (er Error) Error() string {
	var causeStr string
	if er.Err != nil {
		causeStr = fmt.Sprintf(": %s", er.Err.Error())
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("http status = %v, publicMsg = %s", er.Status, er.Description))
	if er.Code != "" {
		sb.WriteString(fmt.Sprintf(", code = %s", er.Code))
	}
	if len(er.Fields) > 0 {
		sb.WriteString(fmt.Sprintf(", fields = %v", er.Fields))
	}
	if causeStr != "" {
		sb.WriteString(causeStr)
	}

	return sb.String()
}

func (er Error) GetData() (int, string, string, []FieldError) {
	return er.Status, er.Code, er.Description, er.Fields
}

func FromFieldErrors(fieldErrs []FieldError) error {
	return Error{
		Status:      http.StatusBadRequest,
		Code:        "invalid_request",
		Description: "Invalid request",
		Fields:      fieldErrs,
	}
}

// FieldErrsFromValidateErrs provides a mapping from github.com/go-playground/validator errors to fieldErrors
func FieldErrsFromValidateErrs(errors validator.ValidationErrors) []FieldError {
	var fieldErrs []FieldError
	var desc string
	for _, err := range errors {
		desc = "Invalid field, failed validation check: " + err.Tag()
		fieldErrs = append(fieldErrs, FieldError{
			Field:       err.Field(),
			Description: desc,
		})
	}
	return fieldErrs
}

// For use when dealing with external clients
// hides any internal information that should not be exposed publicly
var (
	// Err400Default ...
	Err400Default = &Error{
		Status:      http.StatusBadRequest,
		Description: http.StatusText(http.StatusBadRequest),
		Code:        "generic_bad_request",
	}

	// Err401Default ...
	Err401Default = &Error{
		Status:      http.StatusUnauthorized,
		Description: http.StatusText(http.StatusUnauthorized),
		Code:        "generic_unauthorized",
	}

	// Err403Default ...
	Err403Default = &Error{
		Status:      http.StatusForbidden,
		Description: http.StatusText(http.StatusForbidden),
		Code:        "generic_forbidden",
	}

	// Err404Default ...
	Err404Default = &Error{
		Status:      http.StatusNotFound,
		Description: http.StatusText(http.StatusNotFound),
		Code:        "generic_not_found",
	}

	// Err409Default ...
	Err409Default = &Error{
		Status:      http.StatusConflict,
		Description: http.StatusText(http.StatusConflict),
		Code:        "generic_conflict",
	}

	// Err415Default ...
	Err415Default = &Error{
		Status:      http.StatusUnsupportedMediaType,
		Description: http.StatusText(http.StatusUnsupportedMediaType),
		Code:        "generic_unsupported_media_type",
	}

	// Err422Default ...
	Err422Default = &Error{
		Status:      http.StatusUnprocessableEntity,
		Description: "Your request has not been processed, some precondition failed",
		Code:        "generic_unprocessable_entity",
	}

	// Err429Default ...
	Err429Default = &Error{
		Status:      http.StatusTooManyRequests,
		Description: http.StatusText(http.StatusTooManyRequests),
		Code:        "generic_too_many_requests",
	}

	// Err499Default ...
	Err499Default = &Error{
		Status:      499,
		Description: "Request Cancelled",
		Code:        "generic_request_cancelled",
	}

	// Err500Default ...
	Err500Default = &Error{
		Status:      http.StatusInternalServerError,
		Description: http.StatusText(http.StatusInternalServerError),
		Code:        "generic_internal_server_error",
	}

	// Err504Default ...
	Err504Default = &Error{
		Status:      http.StatusGatewayTimeout,
		Description: http.StatusText(http.StatusGatewayTimeout),
		Code:        "generic_gateway_timeout",
	}
)
