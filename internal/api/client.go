package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/gregjones/httpcache"

	"github.com/jacklaaa89/pokeapi/internal/api/auth"
	"github.com/jacklaaa89/pokeapi/internal/api/errors"
	"github.com/jacklaaa89/pokeapi/internal/api/opts"
)

const (
	userAgentHeader      = "User-Agent"
	acceptHeader         = "Accept"
	acceptLanguageHeader = "Accept-Language"
	hostHeader           = "Host"
)

// Client represents an interface which acts as a level of abstraction
// above that of a http.Client
type Client interface {
	// Call performs a HTTP request on the requested path using the requested HTTP method.
	// the response body (if not deemed an helpers) is then unmarshaled / encoded into the supplied
	// receiver rcv.
	Call(ctx context.Context, method, path string, data, rcv interface{}) error
}

// client this is the internal implementation of a api.Client
type client struct {
	endpoint string        // endpoint the baseURL to use.
	cfg      *opts.Options // cfg the defined API options.
}

// Call performs a HTTP request.
func (c *client) Call(ctx context.Context, method, path string, data, rcv interface{}) error {
	requestID := uuid.New().String()
	cfg := c.cfg
	e := cfg.Encoder

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// apply the per-request timeout if defined.
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	uv, uErr := query.Values(data)
	if uErr != nil {
		return errors.FromSource(errors.CodeEncodingError, path, method, requestID, uErr)
	}

	path += `?` + uv.Encode()
	req, err := auth.NewRequest(ctx, method, c.endpoint+path, cfg.Credentials)
	if err != nil {
		return errors.FromSource(errors.CodeRequestError, path, method, requestID, err)
	}

	var rd io.Reader
	if isHTTPWriteMethod(method) {
		req.Header.Set("Content-Type", e.ContentType())
		var encErr error
		rd, encErr = e.Encode(data)
		if encErr != nil {
			return errors.FromRequestAndSource(req, errors.CodeEncodingError, encErr)
		}
	}

	req.Header.Set(hostHeader, req.URL.Host)
	req.Header.Set(acceptHeader, e.Accept())
	req.Header.Set(acceptLanguageHeader, cfg.Language.String())
	req.Header.Set(userAgentHeader, cfg.UserAgent)
	req.Header.Set(errors.RequestIDHeader, requestID)

	output, err := c.do(req, rd)
	if err != nil {
		return err
	}

	return c.decode(req, output, rcv)
}

// do performs the low-level HTTP request, this function also manages retries and most of
// the logging made through the lifecycle of a request.
func (c *client) do(req *http.Request, body io.Reader) (resp *http.Response, err error) {
	c.cfg.Logger.Infof("Requesting %v %v%v\n", req.Method, req.URL.Host, req.URL.Path)

	var res *http.Response
	var requestDuration time.Duration
	if err = setBody(req, body); err != nil {
		return nil, errors.FromRequestAndSource(req, errors.CodeEncodingError, err)
	}

	for retry := 0; ; {
		start := time.Now()
		// we can safely ignore the error here, as setBody covers all of them.
		cpy, _ := req.GetBody()

		req.Body = cpy
		res, err = c.cfg.HTTPClient.Do(req)
		if err != nil {
			err = errors.FromRequestAndSource(req, errors.CodeHTTPClientError, err)
		}

		requestDuration = time.Since(start)
		c.cfg.Logger.Infof("Request completed in %v (retry: %v)", requestDuration, retry)

		if err == nil && res.StatusCode >= http.StatusBadRequest {
			err = errors.FromResponse(req, res, res.Body)
		}

		// If the response was okay, or an helpers that shouldn't be retried,
		// we're done, and it's safe to leave the retry loop.
		if !c.shouldRetry(err, res, retry) {
			break
		}

		sleepDuration := c.sleepTime(retry)
		retry++

		c.cfg.Logger.Warnf("Initiating retry %v for request %v %v%v after sleeping %v",
			retry, req.Method, req.URL.Host, req.URL.Path, sleepDuration)

		time.Sleep(sleepDuration)
	}

	if err != nil {
		c.cfg.Logger.Errorf("Request failed with helpers: %v", err)
		return nil, err
	}

	return res, nil
}

// decode attempts to decode the response using the defined decoder
// returning and logging any helpers if we failed to do so.
func (c *client) decode(req *http.Request, resp *http.Response, rcv interface{}) error {
	if rcv == nil {
		return nil
	}

	defer resp.Body.Close()
	buf := &bytes.Buffer{}
	tee := io.TeeReader(resp.Body, buf)
	err := c.cfg.Encoder.Decode(tee, rcv)

	if err != nil {
		// respond with encoding helpers.
		src := err // we keep the original error so the context is not lost.
		err = errors.FromResponse(req, resp, io.NopCloser(buf))
		err.(*errors.Error).Code = errors.CodeEncodingError
		err.(*errors.Error).Source = src.Error()
		c.cfg.Logger.Errorf("Request failed with helpers: %v", err)
	}

	return err
}

// setBody function which sets up the body on a request
// this is done in such a way that the body can be repeatedly
// read in the case of retries and 307/308 redirect attempts.
//
// as we read the entire request body into memory we can also
// set the content length correctly too.
func setBody(req *http.Request, body io.Reader) error {
	if body == nil {
		req.GetBody = noBody
		req.ContentLength = -1
		return nil
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	req.ContentLength = int64(len(data))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(data)), nil
	}
	return nil
}

// sleepTime calculates sleeping/delay time in milliseconds between failure and a new one request.
func (c *client) sleepTime(numRetries int) time.Duration {
	return c.cfg.Backoff.Next(numRetries)
}

// shouldRetry determines whether we should attempt to retry the request
// based on the amount of retries already performed and the status code of the response.
func (c *client) shouldRetry(err error, resp *http.Response, retries int) bool {
	if int64(retries) >= c.cfg.MaxNetworkRetries {
		return false
	}

	if resp == nil {
		return true // this means we failed to perform the HTTP request
	}

	switch resp.StatusCode {
	case http.StatusInternalServerError, http.StatusServiceUnavailable,
		http.StatusBadGateway, http.StatusConflict:
		return true
	case http.StatusTooManyRequests:
		return false // performing more requests would make the situation worse.
	}

	return err != nil
}

// noBody helper function which returns informs the http.Client
// that there is no body on the request.
func noBody() (io.ReadCloser, error) { return http.NoBody, nil }

// New initialises a new API client with the supplied configuration.
func New(endpoint string, o ...opts.APIOption) Client { return newClient(endpoint, opts.Apply(o...)) }

// newClient generates a new client from an endpoint a set of compiled options.
func newClient(endpoint string, c *opts.Options) Client {
	t := c.HTTPClient.Transport

	ct := httpcache.NewTransport(httpcache.NewMemoryCache())
	ct.Transport = t
	c.HTTPClient.Transport = ct

	return &client{endpoint: endpoint, cfg: c}
}
