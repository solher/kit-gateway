package library

import (
	"encoding/json"
	"errors"
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

type Handlers struct {
	CreateDocumentHandler      http.Handler
	FindDocumentsHandler       http.Handler
	FindDocumentsByIDHandler   http.Handler
	ReplaceDocumentByIDHandler http.Handler
	DeleteDocumentsByIDHandler http.Handler
}

func MakeHTTPHandlers(ctx context.Context, e client.Endpoints, tracer stdopentracing.Tracer, logger log.Logger) Handlers {
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(common.ServerErrorEncoder),
	}

	createDocumentHandler := httptransport.NewServer(
		ctx,
		e.CreateDocumentEndpoint,
		decodeHTTPCreateDocumentRequest,
		encodeHTTPCreateDocumentResponse,
		append(
			opts,
			httptransport.ServerBefore(
				opentracing.FromHTTPRequest(tracer, "CreateDocument", logger),
				common.AddHTTPAnnotations,
			),
		)...,
	)
	findDocumentsHandler := httptransport.NewServer(
		ctx,
		e.FindDocumentsEndpoint,
		decodeHTTPFindDocumentsRequest,
		encodeHTTPFindDocumentsResponse,
		append(
			opts,
			httptransport.ServerBefore(
				opentracing.FromHTTPRequest(tracer, "FindDocuments", logger),
				common.AddHTTPAnnotations,
			),
		)...,
	)
	findDocumentsByIDHandler := httptransport.NewServer(
		ctx,
		e.FindDocumentsByIDEndpoint,
		decodeHTTPFindDocumentsByIDRequest,
		encodeHTTPFindDocumentsByIDResponse,
		append(
			opts,
			httptransport.ServerBefore(
				opentracing.FromHTTPRequest(tracer, "FindDocumentsByID", logger),
				common.AddHTTPAnnotations,
			),
		)...,
	)
	replaceDocumentByIDHandler := httptransport.NewServer(
		ctx,
		e.ReplaceDocumentByIDEndpoint,
		decodeHTTPReplaceDocumentByIDRequest,
		encodeHTTPReplaceDocumentByIDResponse,
		append(
			opts,
			httptransport.ServerBefore(
				opentracing.FromHTTPRequest(tracer, "ReplaceDocumentByID", logger),
				common.AddHTTPAnnotations,
			),
		)...,
	)
	deleteDocumentsByIDHandler := httptransport.NewServer(
		ctx,
		e.DeleteDocumentsByIDEndpoint,
		decodeHTTPDeleteDocumentsByIDRequest,
		encodeHTTPDeleteDocumentsByIDResponse,
		append(
			opts,
			httptransport.ServerBefore(
				opentracing.FromHTTPRequest(tracer, "DeleteDocumentsByID", logger),
				common.AddHTTPAnnotations,
			),
		)...,
	)
	return Handlers{
		CreateDocumentHandler:      createDocumentHandler,
		FindDocumentsHandler:       findDocumentsHandler,
		FindDocumentsByIDHandler:   findDocumentsByIDHandler,
		ReplaceDocumentByIDHandler: replaceDocumentByIDHandler,
		DeleteDocumentsByIDHandler: deleteDocumentsByIDHandler,
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
		return common.EncodeHTTPError(ctx, w, errors.New(res.Err))
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
		return common.EncodeHTTPError(ctx, w, errors.New(res.Err))
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
		return common.EncodeHTTPError(ctx, w, errors.New(res.Err))
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
		return common.EncodeHTTPError(ctx, w, errors.New(res.Err))
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
		return common.EncodeHTTPError(ctx, w, errors.New(res.Err))
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
