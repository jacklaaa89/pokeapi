package errors

import (
	"encoding/json"
	"io"
	"net/http"
)

// RequestIDHeader the header to send the unique request id
const RequestIDHeader = "X-Request-ID"
const bodySampleSize = 500

// Error a formatted API helpers
type Error struct {
	// Code the helpers code defined
	Code Code `json:"code"`
	// StatusCode the HTTP status code (if applicable) from the API response
	StatusCode int `json:"status_code"`
	// Method is the HTTP method
	Method string `json:"method"`
	// Resource the requested resource path
	Resource string `json:"resource"`
	// RequestID the uuid of the request.
	RequestID string `json:"request_id"`
	// Request a sample of the request body.
	// this will be nil if there was either no body on the request
	// or the error is not associated with performing the API request.
	Request []byte `json:"request,omitempty"`
	// Response a sample of the response body.
	Response []byte `json:"response"`
	// Source this is the underlined error if applicable.
	Source string `json:"source,omitempty"`
}

// Error implements helpers interface.
func (e *Error) Error() string {
	d, _ := json.Marshal(e)
	return string(d)
}

// code default helpers handler which maps HTTP status codes
// to an helpers code which gives a little bit more context to an helpers.
func code(statusCode int) Code {
	switch statusCode {
	case http.StatusBadRequest:
		return CodeInvalidRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusMethodNotAllowed:
		return CodeInvalidOperation
	case http.StatusConflict:
		return CodeConflict
	case http.StatusPreconditionFailed:
		return CodeInvalidContentType
	case http.StatusUnprocessableEntity:
		return CodeValidationError
	case http.StatusTooManyRequests:
		return CodeRateLimitExceeded
	case http.StatusInternalServerError:
		return CodeServerError
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return CodeServerUnavailable
	}

	return CodeUnknownError
}

// FromSource generates an wrapped error from a source error.
func FromSource(code Code, path, method, id string, src error) *Error {
	return &Error{
		Code:      code,
		Method:    method,
		Resource:  path,
		RequestID: id,
		Source:    errorMessage(src),
	}
}

// FromRequestAndSource generates a wrapped error from a HTTP request, overriding the code and source error.
func FromRequestAndSource(req *http.Request, code Code, src error) *Error {
	err := FromRequest(req)
	err.Code, err.Source = code, errorMessage(src)
	return err
}

// FromResponse generates an Error from a request and response.
// The response body is supplied separately as the stream retrieved in the response
// is a one-time read, providing it separately ensures we account for this and always
// return a usable reader.
func FromResponse(req *http.Request, resp *http.Response, r io.ReadCloser) *Error {
	err := FromRequest(req)
	err.Code, err.StatusCode = code(resp.StatusCode), resp.StatusCode
	err.Response = RetrieveSample(r)
	return err
}

// FromRequest generates an error from a http.Request.
// useful when an error occurs before the Response is generated.
func FromRequest(req *http.Request) *Error {
	var b io.ReadCloser = http.NoBody
	if req.GetBody != nil {
		b, _ = req.GetBody()
	}

	return &Error{
		Method:    req.Method,
		Resource:  req.URL.Path,
		RequestID: req.Header.Get(RequestIDHeader),
		Request:   RetrieveSample(b),
	}
}

// RetrieveSample retrieves a sample from a reader
func RetrieveSample(r io.ReadCloser) []byte {
	if r == nil || r == http.NoBody {
		return nil
	}

	defer r.Close()
	d, err := io.ReadAll(r)
	if err != nil {
		return nil
	}

	if len(d) < bodySampleSize {
		return d
	}

	dots := []byte("...")
	return append(d[:bodySampleSize], dots...)
}

// errorMessage attempts to retrieve the error message from an error.
func errorMessage(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
