package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/solher/kit-crud/client"
	"github.com/solher/kit-crud/pb"
	"github.com/solher/kit-gateway/common"

	"golang.org/x/net/context"
)

func MakeHTTPHandler(ctx context.Context, e client.Endpoints, tracer stdopentracing.Tracer, logger log.Logger) http.Handler {
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(common.EncodeHTTPError),
	}

	createDocumentHandler := httptransport.NewServer(
		ctx,
		e.CreateDocumentEndpoint,
		decodeHTTPCreateDocumentRequest,
		encodeHTTPCreateDocumentResponse,
		append(opts, httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "CreateDocument", logger)))...,
	)
	findDocumentsHandler := httptransport.NewServer(
		ctx,
		e.FindDocumentsEndpoint,
		decodeHTTPFindDocumentsRequest,
		encodeHTTPFindDocumentsResponse,
		append(
			opts,
			httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "FindDocuments", logger)),
			httptransport.ServerBefore(opentracing.ToHTTPRequest(tracer, logger)),
			httptransport.ServerBefore(AddHTTPAnnotations(tracer)),
		)...,
	)
	findDocumentsByIDHandler := httptransport.NewServer(
		ctx,
		e.FindDocumentsByIDEndpoint,
		decodeHTTPFindDocumentsByIDRequest,
		encodeHTTPFindDocumentsByIDResponse,
		append(opts, httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "FindDocumentsByID", logger)))...,
	)
	replaceDocumentByIDHandler := httptransport.NewServer(
		ctx,
		e.ReplaceDocumentByIDEndpoint,
		decodeHTTPReplaceDocumentByIDRequest,
		encodeHTTPReplaceDocumentByIDResponse,
		append(opts, httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "ReplaceDocumentByID", logger)))...,
	)
	deleteDocumentsByIDHandler := httptransport.NewServer(
		ctx,
		e.DeleteDocumentsByIDEndpoint,
		decodeHTTPDeleteDocumentsByIDRequest,
		encodeHTTPDeleteDocumentsByIDResponse,
		append(opts, httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "DeleteDocumentsByID", logger)))...,
	)

	r := bone.New()
	r.Post("/documents", createDocumentHandler)
	r.Get("/documents", findDocumentsHandler)
	r.Get("/documents/:ids", findDocumentsByIDHandler)
	r.Put("/documents/:id", replaceDocumentByIDHandler)
	r.Delete("/documents/:ids", deleteDocumentsByIDHandler)
	return r
}

func AddHTTPAnnotations(tracer stdopentracing.Tracer) httptransport.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		span := stdopentracing.SpanFromContext(ctx)
		// if span == nil {
		// 	fmt.Println("nil span")
		// 	return ctx
		// }
		if span == nil {
			// All we can do is create a new root span.
			fmt.Println("nil span")
			span = tracer.StartSpan("")
		}

		span = span.SetTag("foo", "bar")
		span.LogEventWithPayload("url", r.URL.RequestURI())
		return stdopentracing.ContextWithSpan(ctx, span)
	}
}

func decodeHTTPCreateDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var document *pb.Document
	if err := json.NewDecoder(r.Body).Decode(document); err != nil {
		return nil, err
	}
	req := pb.CreateDocumentRequest{
		UserId:   "admin",
		Document: document,
	}
	return req, nil
}

func encodeHTTPCreateDocumentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(*pb.CreateDocumentReply)
	if len(res.Err) > 0 {
		return errors.New(res.Err)
	}
	return common.EncodeHTTPResponse(ctx, w, http.StatusCreated, documentEncoder, res.Document)
}

func decodeHTTPFindDocumentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := &pb.FindDocumentsRequest{
		UserId: "admin",
	}
	return req, nil
}

func encodeHTTPFindDocumentsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(*pb.FindDocumentsReply)
	if len(res.Err) > 0 {
		return errors.New(res.Err)
	}
	return common.EncodeHTTPResponse(ctx, w, http.StatusOK, documentsEncoder, res.Documents)
}

func decodeHTTPFindDocumentsByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	val := bone.GetValue(r, ":ids")
	ids := strings.Split(val, ",")
	req := &pb.FindDocumentsByIdRequest{
		UserId: "admin",
		Ids:    ids,
	}
	return req, nil
}

func encodeHTTPFindDocumentsByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(*pb.FindDocumentsByIdReply)
	if len(res.Err) > 0 {
		return errors.New(res.Err)
	}
	if len(res.Documents) == 1 {
		return common.EncodeHTTPResponse(ctx, w, http.StatusOK, documentEncoder, res.Documents[0])
	}
	return common.EncodeHTTPResponse(ctx, w, http.StatusOK, documentsEncoder, res.Documents)
}

func decodeHTTPReplaceDocumentByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var document *pb.Document
	if err := json.NewDecoder(r.Body).Decode(document); err != nil {
		return nil, err
	}
	req := &pb.ReplaceDocumentByIdRequest{
		UserId:   "admin",
		Id:       bone.GetValue(r, ":id"),
		Document: document,
	}
	return req, nil
}

func encodeHTTPReplaceDocumentByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(*pb.ReplaceDocumentByIdReply)
	if len(res.Err) > 0 {
		return errors.New(res.Err)
	}
	return common.EncodeHTTPResponse(ctx, w, http.StatusOK, documentEncoder, res.Document)
}

func decodeHTTPDeleteDocumentsByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	val := bone.GetValue(r, ":ids")
	ids := strings.Split(val, ",")
	req := &pb.DeleteDocumentsByIdRequest{
		UserId: "admin",
		Ids:    ids,
	}
	return req, nil
}

func encodeHTTPDeleteDocumentsByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(*pb.DeleteDocumentsByIdReply)
	if len(res.Err) > 0 {
		return errors.New(res.Err)
	}
	if len(res.Documents) == 1 {
		return common.EncodeHTTPResponse(ctx, w, http.StatusOK, documentEncoder, res.Documents[0])
	}
	return common.EncodeHTTPResponse(ctx, w, http.StatusOK, documentsEncoder, res.Documents)
}

func documentEncoder(w http.ResponseWriter, response interface{}) error {
	document := response.(*pb.Document)
	return documentsEncoder(w, []*pb.Document{document})
}

func documentsEncoder(w http.ResponseWriter, response interface{}) error {
	documents := response.([]*pb.Document)

	output := make([]struct {
		ID      *string `json:"id"`
		UserID  *string `json:"userId"`
		Content *string `json:"content"`
	}, len(documents))

	for i := range documents {
		output[i].ID = &documents[i].Id
		output[i].UserID = &documents[i].UserId
		output[i].Content = &documents[i].Content
	}

	return json.NewEncoder(w).Encode(output)
}
