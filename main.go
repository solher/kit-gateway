package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
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
		zipkinAddr = flag.String("zipkin.addr", "", "Enable Zipkin tracing via a Kafka server host:port")
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
			collector, err := zipkin.NewKafkaCollector(
				strings.Split(*zipkinAddr, ","),
				zipkin.KafkaLogger(logger),
			)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			tracer, err = zipkin.NewTracer(
				zipkin.NewRecorder(collector, false, "localhost:80", "Library"),
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
		libraryService, err = libraryClient.NewGRPC(*crudAddr, logger)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
	}

	// Endpoint domain.
	var createDocumentEndpoint endpoint.Endpoint
	{
		createDocumentEndpoint = libraryClient.MakeCreateDocumentEndpoint(libraryService)
		createDocumentEndpoint = opentracing.TraceServer(tracer, "CreateDocument")(createDocumentEndpoint)
		// createDocumentEndpoint = EndpointLoggingMiddleware(logger)(createDocumentEndpoint)
	}
	var findDocumentsEndpoint endpoint.Endpoint
	{
		findDocumentsEndpoint = libraryClient.MakeFindDocumentsEndpoint(libraryService)
		findDocumentsEndpoint = opentracing.TraceServer(tracer, "FindDocuments")(findDocumentsEndpoint)
		// findDocumentsEndpoint = EndpointLoggingMiddleware(logger)(findDocumentsEndpoint)
	}
	var findDocumentsByIDEndpoint endpoint.Endpoint
	{
		findDocumentsByIDEndpoint = libraryClient.MakeFindDocumentsByIDEndpoint(libraryService)
		findDocumentsByIDEndpoint = opentracing.TraceServer(tracer, "FindDocumentsByID")(findDocumentsByIDEndpoint)
		// findDocumentsByIDEndpoint = EndpointLoggingMiddleware(logger)(findDocumentsByIDEndpoint)
	}
	var replaceDocumentByIDEndpoint endpoint.Endpoint
	{
		replaceDocumentByIDEndpoint = libraryClient.MakeReplaceDocumentByIDEndpoint(libraryService)
		replaceDocumentByIDEndpoint = opentracing.TraceServer(tracer, "ReplaceDocumentByID")(replaceDocumentByIDEndpoint)
		// replaceDocumentByIDEndpoint = EndpointLoggingMiddleware(logger)(replaceDocumentByIDEndpoint)
	}
	var deleteDocumentsByIDEndpoint endpoint.Endpoint
	{
		deleteDocumentsByIDEndpoint = libraryClient.MakeDeleteDocumentsByIDEndpoint(libraryService)
		deleteDocumentsByIDEndpoint = opentracing.TraceServer(tracer, "DeleteDocumentsByID")(deleteDocumentsByIDEndpoint)
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

	logger.Log("msg", "listening on "+*httpAddr+" (HTTP)")

	if err := http.ListenAndServe(*httpAddr, handler); err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
}
