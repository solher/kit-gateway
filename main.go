package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-zoo/bone"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"

	libraryClient "github.com/solher/kit-crud/client"

	"github.com/solher/kit-gateway/library"
	"golang.org/x/net/context"
)

func main() {
	var (
		httpAddr   = flag.String("http.addr", ":3000", "Address for HTTP server")
		zipkinAddr = flag.String("zipkin.addr", "", "Enable Zipkin tracing via a Scribe server host:port")
		crudAddr   = flag.String("crud.addr", "", "Address for kit-crud")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	// Tracing domain.
	var tracer stdopentracing.Tracer
	{
		if *zipkinAddr != "" {
			logger := log.NewContext(logger).With("tracer", "Zipkin")
			logger.Log("msg", "sending trace to "+*zipkinAddr)
			collector, err := zipkin.NewScribeCollector(
				*zipkinAddr,
				3*time.Second,
				zipkin.ScribeLogger(logger),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "kit-gateway:3000", "Gateway"),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
		} else {
			logger := log.NewContext(logger).With("tracer", "none")
			logger.Log("msg", "tracing disabled")
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	// Business domain.
	var libraryService libraryClient.Service
	{
		var err error
		libraryService, err = libraryClient.NewGRPC(*crudAddr, tracer, logger)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
	}

	// Endpoint domain.
	var createDocumentEndpoint endpoint.Endpoint
	{
		createDocumentEndpoint = libraryClient.MakeCreateDocumentEndpoint(libraryService)
	}
	var findDocumentsEndpoint endpoint.Endpoint
	{
		findDocumentsEndpoint = libraryClient.MakeFindDocumentsEndpoint(libraryService)
	}
	var findDocumentsByIDEndpoint endpoint.Endpoint
	{
		findDocumentsByIDEndpoint = libraryClient.MakeFindDocumentsByIDEndpoint(libraryService)
	}
	var replaceDocumentByIDEndpoint endpoint.Endpoint
	{
		replaceDocumentByIDEndpoint = libraryClient.MakeReplaceDocumentByIDEndpoint(libraryService)
	}
	var deleteDocumentsByIDEndpoint endpoint.Endpoint
	{
		deleteDocumentsByIDEndpoint = libraryClient.MakeDeleteDocumentsByIDEndpoint(libraryService)
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

	libraryHandlers := library.MakeHTTPHandlers(ctx, libraryEndpoints, tracer, logger)

	r := bone.New()
	r.Post("/library/documents", libraryHandlers.CreateDocumentHandler)
	r.Get("/library/documents", libraryHandlers.FindDocumentsHandler)
	r.Get("/library/documents/:ids", libraryHandlers.FindDocumentsByIDHandler)
	r.Put("/library/documents/:id", libraryHandlers.ReplaceDocumentByIDHandler)
	r.Delete("/library/documents/:ids", libraryHandlers.DeleteDocumentsByIDHandler)

	handler := HTTPLoggingMiddleware(logger)(r)

	logger.Log("msg", "listening on "+*httpAddr+" (HTTP)")

	if err := http.ListenAndServe(*httpAddr, handler); err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
}
