package common

import (
	"encoding/json"
	"net/http"

	"github.com/solher/kit-crud/library"
	"golang.org/x/net/context"
)

type Encoder func(w http.ResponseWriter, response interface{}) error

func EncodeHTTPResponse(ctx context.Context, w http.ResponseWriter, code int, encoder Encoder, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return encoder(w, response)
}

func EncodeHTTPError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var apiErr APIError
	switch err {
	case library.ErrNotFound:
		apiErr = ErrForbidden
	default:
		apiErr = ErrInternal
	}
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}
