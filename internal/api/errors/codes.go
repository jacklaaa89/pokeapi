package errors

type Code string

const (
	CodeUnknownError Code = "unknown_error"

	CodeInvalidRequest     Code = "invalid_request"
	CodeForbidden          Code = "forbidden"
	CodeUnauthorized       Code = "unauthorized"
	CodeNotFound           Code = "not_found"
	CodeInvalidOperation   Code = "invalid_operation"
	CodeConflict           Code = "resource_conflict"
	CodeInvalidContentType Code = "invalid_content_type"
	CodeValidationError    Code = "validation_error"
	CodeRateLimitExceeded  Code = "rate_limit_exceeded"
	CodeServerError        Code = "server_error"
	CodeServerUnavailable  Code = "server_unavailable"

	CodeEncodingError   Code = "encoding_error"
	CodeRequestError    Code = "request_error"
	CodeHTTPClientError Code = "http_client_error"
)
