package common

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/solher/kit-crud/library"
	"golang.org/x/net/context"
)

func AddHTTPAnnotations(ctx context.Context, r *http.Request) context.Context {
	span := stdopentracing.SpanFromContext(ctx)
	if span == nil {
		return ctx
	}
	buf := bytes.NewBuffer(nil)
	body, _ := ioutil.ReadAll(io.TeeReader(r.Body, buf))
	r.Body = ioutil.NopCloser(buf)
	span = span.SetTag("req.body", string(body))
	span = span.SetTag("req.method", r.Method)
	span = span.SetTag("req.url", r.URL.String())
	span = span.SetTag("req.remote", r.RemoteAddr)
	span = span.SetTag("req.agent", r.UserAgent())
	return stdopentracing.ContextWithSpan(ctx, span)
}

type Encoder func(w http.ResponseWriter, response interface{}) error

func EncodeHTTPResponse(ctx context.Context, w http.ResponseWriter, status int, encoder Encoder, response interface{}) error {
	defer func() {
		span := stdopentracing.SpanFromContext(ctx)
		if span != nil {
			span = span.SetTag("res.status", status)
			span.Finish()
		}
	}()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return encoder(w, response)
}

func EncodeHTTPError(ctx context.Context, err error, w http.ResponseWriter) {
	var apiErr APIError
	switch err {
	case library.ErrNotFound:
		apiErr = ErrForbidden
	default:
		apiErr = ErrInternal
	}
	defer func() {
		span := stdopentracing.SpanFromContext(ctx)
		if span != nil {
			span = span.SetTag("error", err.Error())
			span = span.SetTag("res.status", apiErr.Status)
			span = span.SetTag("res.description", apiErr.Description)
			span = span.SetTag("res.errorCode", apiErr.ErrorCode)
			span = span.SetTag("res.params", apiErr.Params)
			span.Finish()
		}
	}()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}
