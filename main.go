package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-zoo/bone"
	libraryClient "github.com/solher/kit-crud/client"
	"github.com/solher/kit-gateway/library"
	"golang.org/x/net/context"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":3000", "Address for HTTP server")
		crudAddr = flag.String("crud.addr", "localhost:8082", "Address for kit-crud")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	// Business domain.
	var libraryService libraryClient.Service
	{
		var err error
		libraryService, err = libraryClient.NewGRPC(*crudAddr, logger)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	// Endpoint domain.
	var createDocumentEndpoint endpoint.Endpoint
	{
		createDocumentEndpoint = libraryClient.MakeCreateDocumentEndpoint(libraryService)
		// createDocumentEndpoint = EndpointLoggingMiddleware(logger)(createDocumentEndpoint)
	}
	var findDocumentsEndpoint endpoint.Endpoint
	{
		findDocumentsEndpoint = libraryClient.MakeFindDocumentsEndpoint(libraryService)
		// findDocumentsEndpoint = EndpointLoggingMiddleware(logger)(findDocumentsEndpoint)
	}
	var findDocumentsByIDEndpoint endpoint.Endpoint
	{
		findDocumentsByIDEndpoint = libraryClient.MakeFindDocumentsByIDEndpoint(libraryService)
		// findDocumentsByIDEndpoint = EndpointLoggingMiddleware(logger)(findDocumentsByIDEndpoint)
	}
	var replaceDocumentByIDEndpoint endpoint.Endpoint
	{
		replaceDocumentByIDEndpoint = libraryClient.MakeReplaceDocumentByIDEndpoint(libraryService)
		// replaceDocumentByIDEndpoint = EndpointLoggingMiddleware(logger)(replaceDocumentByIDEndpoint)
	}
	var deleteDocumentsByIDEndpoint endpoint.Endpoint
	{
		deleteDocumentsByIDEndpoint = libraryClient.MakeDeleteDocumentsByIDEndpoint(libraryService)
		// deleteDocumentsByIDEndpoint = EndpointLoggingMiddleware(logger)(deleteDocumentsByIDEndpoint)
	}

	libraryEndpoints := libraryClient.Endpoints{
		CreateDocumentEndpoint:      createDocumentEndpoint,
		FindDocumentsEndpoint:       findDocumentsEndpoint,
		FindDocumentsByIDEndpoint:   findDocumentsByIDEndpoint,
		ReplaceDocumentByIDEndpoint: replaceDocumentByIDEndpoint,
		DeleteDocumentsByIDEndpoint: deleteDocumentsByIDEndpoint,
	}

	// Transport domain.
	ctx := context.Background()

	r := bone.New()
	r.SubRoute("/library", library.MakeHTTPHandler(ctx, libraryEndpoints, logger))

	handler := HTTPLoggingMiddleware(logger)(r)

	if err := http.ListenAndServe(*httpAddr, handler); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
