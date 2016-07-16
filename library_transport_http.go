package main

import (
	"encoding/json"
	"net/http"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/solher/kit-crud/client"

	"golang.org/x/net/context"
)

func MakeLibraryHTTPHandler(ctx context.Context, e client.Endpoints) http.Handler {
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeHTTPError),
	}

	createDocumentHandler := httptransport.NewServer(
		ctx,
		e.CreateDocumentEndpoint,
		decodeHTTPCreateDocumentRequest,
		encodeHTTPCreateDocumentResponse,
		opts...,
	)
	findDocumentsHandler := httptransport.NewServer(
		ctx,
		e.FindDocumentsEndpoint,
		decodeHTTPFindDocumentsRequest,
		encodeHTTPFindDocumentsResponse,
		opts...,
	)
	findDocumentsByIDHandler := httptransport.NewServer(
		ctx,
		e.FindDocumentsByIDEndpoint,
		decodeHTTPFindDocumentsByIDRequest,
		encodeHTTPFindDocumentsByIDResponse,
		opts...,
	)
	replaceDocumentByIDHandler := httptransport.NewServer(
		ctx,
		e.ReplaceDocumentByIDEndpoint,
		decodeHTTPReplaceDocumentByIDRequest,
		encodeHTTPReplaceDocumentByIDResponse,
		opts...,
	)
	deleteDocumentsByIDHandler := httptransport.NewServer(
		ctx,
		e.DeleteDocumentsByIDEndpoint,
		decodeHTTPDeleteDocumentsByIDRequest,
		encodeHTTPDeleteDocumentsByIDResponse,
		opts...,
	)

	r := bone.New()
	r.Post("/documents", createDocumentHandler)
	r.Get("/documents", findDocumentsHandler)
	r.Get("/documents/:ids", findDocumentsByIDHandler)
	r.Put("/documents/:id", replaceDocumentByIDHandler)
	r.Delete("/documents/:ids", deleteDocumentsByIDHandler)
	return r
}

func decodeHTTPCreateDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var document *client.Document
	if err := json.NewDecoder(r.Body).Decode(document); err != nil {
		return nil, err
	}
	req := client.CreateDocumentRequest{
		UserID:   "admin",
		Document: document,
	}
	return req, nil
}

func encodeHTTPCreateDocumentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(client.CreateDocumentResponse)
	if res.Err != nil {
		return res.Err
	}
	return encodeHTTPResponse(ctx, w, http.StatusCreated, res.Document)
}

func decodeHTTPFindDocumentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := client.FindDocumentsRequest{
		UserID: "admin",
	}
	return req, nil
}

func encodeHTTPFindDocumentsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(client.FindDocumentsResponse)
	if res.Err != nil {
		return res.Err
	}
	return encodeHTTPResponse(ctx, w, http.StatusOK, res.Documents)
}

func decodeHTTPFindDocumentsByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	val := bone.GetValue(r, ":ids")
	ids := strings.Split(val, ",")
	req := client.FindDocumentsByIDRequest{
		UserID: "admin",
		IDs:    ids,
	}
	return req, nil
}

func encodeHTTPFindDocumentsByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(client.FindDocumentsByIDResponse)
	if res.Err != nil {
		return res.Err
	}
	if len(res.Documents) == 1 {
		return encodeHTTPResponse(ctx, w, http.StatusOK, res.Documents[0])
	}
	return encodeHTTPResponse(ctx, w, http.StatusOK, res.Documents)
}

func decodeHTTPReplaceDocumentByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var document *client.Document
	if err := json.NewDecoder(r.Body).Decode(document); err != nil {
		return nil, err
	}
	req := client.ReplaceDocumentByIDRequest{
		UserID:   "admin",
		ID:       bone.GetValue(r, ":id"),
		Document: document,
	}
	return req, nil
}

func encodeHTTPReplaceDocumentByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(client.ReplaceDocumentByIDResponse)
	if res.Err != nil {
		return res.Err
	}
	return encodeHTTPResponse(ctx, w, http.StatusOK, res.Document)
}

func decodeHTTPDeleteDocumentsByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	val := bone.GetValue(r, ":ids")
	ids := strings.Split(val, ",")
	req := client.DeleteDocumentsByIDRequest{
		UserID: "admin",
		IDs:    ids,
	}
	return req, nil
}

func encodeHTTPDeleteDocumentsByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(client.DeleteDocumentsByIDResponse)
	if res.Err != nil {
		return res.Err
	}
	if len(res.Documents) == 1 {
		return encodeHTTPResponse(ctx, w, http.StatusOK, res.Documents[0])
	}
	return encodeHTTPResponse(ctx, w, http.StatusOK, res.Documents)
}
