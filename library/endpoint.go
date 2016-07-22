package library

// import (
// 	"errors"

// 	"github.com/go-kit/kit/endpoint"
// 	"github.com/solher/kit-crud/client"
// 	"github.com/solher/kit-crud/pb"
// 	"golang.org/x/net/context"
// )

// type errorer interface {
// }

// type createDocumentResponse struct {
// 	pb.CreateDocumentReply
// 	errorer
// }

// func MakeCreateDocumentEndpoint(s client.Service) endpoint.Endpoint {
// 	return func(_ context.Context, request interface{}) (interface{}, error) {
// 		req := request.(*pb.CreateDocumentRequest)
// 		document, err := s.CreateDocument(req.UserId, req.Document)
// 		return &pb.CreateDocumentReply{
// 			Document: document,
// 			Err:      toPBError(err),
// 		}, nil
// 	}
// }

// func MakeFindDocumentsEndpoint(s client.Service) endpoint.Endpoint {
// 	return func(_ context.Context, request interface{}) (interface{}, error) {
// 		req := request.(*pb.FindDocumentsRequest)
// 		documents, err := s.FindDocuments(req.UserId)
// 		return &pb.FindDocumentsReply{
// 			Documents: documents,
// 			Err:       toPBError(err),
// 		}, nil
// 	}
// }

// func MakeFindDocumentsByIDEndpoint(s client.Service) endpoint.Endpoint {
// 	return func(_ context.Context, request interface{}) (interface{}, error) {
// 		req := request.(*pb.FindDocumentsByIdRequest)
// 		documents, err := s.FindDocumentsByID(req.UserId, req.Ids)
// 		return &pb.FindDocumentsByIdReply{
// 			Documents: documents,
// 			Err:       toPBError(err),
// 		}, nil
// 	}
// }

// func MakeReplaceDocumentByIDEndpoint(s client.Service) endpoint.Endpoint {
// 	return func(_ context.Context, request interface{}) (interface{}, error) {
// 		req := request.(*pb.ReplaceDocumentByIdRequest)
// 		document, err := s.ReplaceDocumentByID(req.UserId, req.Id, req.Document)
// 		return &pb.ReplaceDocumentByIdReply{
// 			Document: document,
// 			Err:      toPBError(err),
// 		}, nil
// 	}
// }

// func MakeDeleteDocumentsByIDEndpoint(s client.Service) endpoint.Endpoint {
// 	return func(_ context.Context, request interface{}) (interface{}, error) {
// 		req := request.(*pb.DeleteDocumentsByIdRequest)
// 		documents, err := s.DeleteDocumentsByID(req.UserId, req.Ids)
// 		return &pb.DeleteDocumentsByIdReply{
// 			Documents: documents,
// 			Err:       toPBError(err),
// 		}, nil
// 	}
// }

// func toError(err string) error {
// 	if len(err) == 0 {
// 		return nil
// 	}
// 	return errors.New(err)
// }

// func toPBError(err error) string {
// 	if err == nil {
// 		return ""
// 	}
// 	return err.Error()
// }
