package es

import (
	"context"
	"log"
	"net/http"
	"time"
)

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

// Transport implements http.RoundTripper. When set as Transport of http.Client, it executes HTTP requests with logging.
// No field is mandatory.
type LoggingTransport struct {
	LoggingTransport http.RoundTripper
	LogRequest       func(req *http.Request)
	LogResponse      func(resp *http.Response)
}

// THe default logging transport that wraps http.DefaultTransport.
var DefaultTransport = &LoggingTransport{
	LoggingTransport: http.DefaultTransport,
}

// Used if transport.LogRequest is not set.
var DefaultLogRequest = func(req *http.Request) {
	log.Printf("--> %s %s", req.Method, req.URL)
}

// Used if transport.LogResponse is not set.
var DefaultLogResponse = func(resp *http.Response) {
	ctx := resp.Request.Context()
	if start, ok := ctx.Value(ContextKeyRequestStart).(time.Time); ok {
		log.Printf("<-- %d %s (%s)", resp.StatusCode, resp.Request.URL, time.Since(start))
	} else {
		log.Printf("<-- %d %s", resp.StatusCode, resp.Request.URL)
	}
}

type contextKey struct {
	name string
}

var ContextKeyRequestStart = &contextKey{"RequestStart"}

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP request with request/response logging.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), ContextKeyRequestStart, time.Now())
	req = req.WithContext(ctx)

	t.logRequest(req)

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp)

	return resp, err
}

func (t *LoggingTransport) logRequest(req *http.Request) {
	if t.LogRequest != nil {
		t.LogRequest(req)
	} else {
		DefaultLogRequest(req)
	}
}

func (t *LoggingTransport) logResponse(resp *http.Response) {
	if t.LogResponse != nil {
		t.LogResponse(resp)
	} else {
		DefaultLogResponse(resp)
	}
}

func (t *LoggingTransport) transport() http.RoundTripper {
	if t.LoggingTransport != nil {
		return t.LoggingTransport
	}

	return http.DefaultTransport
}
