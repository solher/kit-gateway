package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/solher/kit-crud/client"
	"golang.org/x/net/context"
)

func encodeHTTPResponse(ctx context.Context, w http.ResponseWriter, code int, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(response)
}

func encodeHTTPError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var apiErr APIError
	switch err {
	case client.ErrNotFound:
		apiErr = ErrForbidden
	default:
		apiErr = ErrInternal
	}
	fmt.Println(err)
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(apiErr)
}
