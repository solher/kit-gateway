package common

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/solher/kit-crud/library"
	"golang.org/x/net/context"
)

func AddHTTPAnnotations(ctx context.Context, r *http.Request) context.Context {
	if span := stdopentracing.SpanFromContext(ctx); span != nil {
		buf := bytes.NewBuffer(nil)
		body, _ := ioutil.ReadAll(io.TeeReader(r.Body, buf))
		r.Body = ioutil.NopCloser(buf)
		span = span.SetTag("transport", "HTTP")
		span = span.SetTag("req.body", string(body))
		span = span.SetTag("req.method", r.Method)
		span = span.SetTag("req.url", r.URL.String())
		span = span.SetTag("req.remote", r.RemoteAddr)
		span = span.SetTag("req.agent", r.UserAgent())
		ctx = stdopentracing.ContextWithSpan(ctx, span)
	}
	return ctx
}

type Encoder func(w http.ResponseWriter, response interface{}) error

func EncodeHTTPResponse(ctx context.Context, w http.ResponseWriter, status int, encoder Encoder, response interface{}) error {
	defer func() {
		if span := stdopentracing.SpanFromContext(ctx); span != nil {
			span = span.SetTag("res.status", status)
			span.Finish()
		}
	}()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return encoder(w, response)
}

func EncodeHTTPError(ctx context.Context, w http.ResponseWriter, err error) error {
	var apiErr APIError
	switch err {
	case library.ErrNotFound:
		apiErr = ErrForbidden
	default:
		apiErr = ErrInternal
	}
	return encodeAPIError(ctx, apiErr, w)
}

func ServerErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	switch e := err.(type) {
	case httptransport.Error:
		switch e.Domain {
		case httptransport.DomainDecode:
			encodeAPIError(ctx, ErrBodyDecoding, w)
			return
		}
	}
	if span := stdopentracing.SpanFromContext(ctx); span != nil {
		span = span.SetTag("error", err)
		ctx = stdopentracing.ContextWithSpan(ctx, span)
	}
	encodeAPIError(ctx, ErrInternal, w)
}

func encodeAPIError(ctx context.Context, err APIError, w http.ResponseWriter) error {
	defer func() {
		if span := stdopentracing.SpanFromContext(ctx); span != nil {
			span = span.SetTag("res.status", err.Status)
			span = span.SetTag("res.description", err.Description)
			span = span.SetTag("res.errorCode", err.ErrorCode)
			span = span.SetTag("res.params", err.Params)
			span.Finish()
		}
	}()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(err.Status)
	return json.NewEncoder(w).Encode(err)
}
