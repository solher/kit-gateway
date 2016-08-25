package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-zoo/bone"
	stdopentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"

	libraryClient "github.com/solher/kit-crud/client"

	"github.com/solher/kit-gateway/library"
	"golang.org/x/net/context"
)

func main() {
	var (
		httpAddr    = flag.String("http.addr", ":3000", "Address for HTTP server")
		appdashAddr = flag.String("appdash.addr", "", "Enable Appdash tracing via server host:port")
		crudAddr    = flag.String("crud.addr", "", "Address for kit-crud")
	)
	flag.Parse()

	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

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
		if *appdashAddr != "" {
			logger := log.NewContext(logger).With("tracer", "Appdash")
			logger.Log("msg", "sending trace to "+*appdashAddr)
			tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashAddr))
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
			exitCode = 1
			return
		}
	}

	// Endpoint domain.
	var createDocumentEndpoint endpoint.Endpoint
	{
		createDocumentEndpoint = libraryClient.MakeCreateDocumentEndpoint(libraryService)
		createDocumentEndpoint = EndpointTracingMiddleware(createDocumentEndpoint)
	}
	var findDocumentsEndpoint endpoint.Endpoint
	{
		findDocumentsEndpoint = libraryClient.MakeFindDocumentsEndpoint(libraryService)
		findDocumentsEndpoint = EndpointTracingMiddleware(findDocumentsEndpoint)
	}
	var findDocumentsByIDEndpoint endpoint.Endpoint
	{
		findDocumentsByIDEndpoint = libraryClient.MakeFindDocumentsByIDEndpoint(libraryService)
		findDocumentsByIDEndpoint = EndpointTracingMiddleware(findDocumentsByIDEndpoint)
	}
	var replaceDocumentByIDEndpoint endpoint.Endpoint
	{
		replaceDocumentByIDEndpoint = libraryClient.MakeReplaceDocumentByIDEndpoint(libraryService)
		replaceDocumentByIDEndpoint = EndpointTracingMiddleware(replaceDocumentByIDEndpoint)
	}
	var deleteDocumentsByIDEndpoint endpoint.Endpoint
	{
		deleteDocumentsByIDEndpoint = libraryClient.MakeDeleteDocumentsByIDEndpoint(libraryService)
		deleteDocumentsByIDEndpoint = EndpointTracingMiddleware(deleteDocumentsByIDEndpoint)
	}

	libraryEndpoints := libraryClient.Endpoints{
		CreateDocumentEndpoint:      createDocumentEndpoint,
		FindDocumentsEndpoint:       findDocumentsEndpoint,
		FindDocumentsByIDEndpoint:   findDocumentsByIDEndpoint,
		ReplaceDocumentByIDEndpoint: replaceDocumentByIDEndpoint,
		DeleteDocumentsByIDEndpoint: deleteDocumentsByIDEndpoint,
	}

	// Mechanical domain.
	ctx := context.Background()
	errc := make(chan error)

	// Transport domain.
	libraryHandlers := library.MakeHTTPHandlers(ctx, libraryEndpoints, tracer, logger)

	r := bone.New()
	r.Post("/library/documents", libraryHandlers.CreateDocumentHandler)
	r.Get("/library/documents", libraryHandlers.FindDocumentsHandler)
	r.Get("/library/documents/:ids", libraryHandlers.FindDocumentsByIDHandler)
	r.Put("/library/documents/:id", libraryHandlers.ReplaceDocumentByIDHandler)
	r.Delete("/library/documents/:ids", libraryHandlers.DeleteDocumentsByIDHandler)

	handler := HTTPLoggingMiddleware(logger)(r)

	conn, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		logger.Log("err", err)
		exitCode = 1
		return
	}
	defer conn.Close()
	logger.Log("msg", "listening on "+*httpAddr+" (HTTP)")
	go func() {
		if err := http.Serve(conn, handler); err != nil {
			errc <- err
			return
		}
	}()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		logger.Log(
			"signal", fmt.Sprintf("%s", <-c),
			"msg", "gracefully shutting down",
		)
		errc <- nil
	}()

	if err := <-errc; err != nil {
		logger.Log("err", err)
		exitCode = 1
	}
}
