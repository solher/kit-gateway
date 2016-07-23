package main

import (
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
)

func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("error", err, "took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)
		}
	}
}

type HTTPMiddleware func(http.Handler) http.Handler

func HTTPLoggingMiddleware(logger log.Logger) HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return &httpLoggingMiddleware{logger: logger, next: next}
	}
}

type httpLoggingMiddleware struct {
	logger log.Logger
	next   http.Handler
}

func (mw *httpLoggingMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	proxy := &writerProxy{ResponseWriter: rw}
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", r.Method,
			"url", r.URL.String(),
			"remote", r.RemoteAddr,
			"agent", r.UserAgent(),
			"code", proxy.code,
			"took", time.Since(begin),
		)
	}(time.Now())
	mw.next.ServeHTTP(proxy, r)
}

// writerProxy holds the status code and a
// flag in addition to http.ResponseWriter
type writerProxy struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
}

// WriteHeader stores the status code and writes header
func (wp *writerProxy) WriteHeader(code int) {
	if !wp.wroteHeader {
		wp.code = code
		wp.wroteHeader = true
		wp.ResponseWriter.WriteHeader(code)
	}
}

// Write writes the bytes and calls MaybeWriteHeader
func (wp *writerProxy) Write(buf []byte) (int, error) {
	if !wp.wroteHeader {
		wp.WriteHeader(http.StatusOK)
	}
	return wp.ResponseWriter.Write(buf)
}
