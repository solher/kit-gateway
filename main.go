package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-zoo/bone"
	libraryClient "github.com/solher/kit-crud/client"
	"golang.org/x/net/context"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":3000", "Address for HTTP server")
	)
	flag.Parse()

	library, err := libraryClient.NewGRPC("")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	libraryEndpoints := libraryClient.Endpoints{}
	libraryEndpoints.CreateDocumentEndpoint = libraryClient.MakeCreateDocumentEndpoint(library)
	libraryEndpoints.FindDocumentsEndpoint = libraryClient.MakeFindDocumentsEndpoint(library)
	libraryEndpoints.FindDocumentsByIDEndpoint = libraryClient.MakeFindDocumentsByIDEndpoint(library)
	libraryEndpoints.ReplaceDocumentByIDEndpoint = libraryClient.MakeReplaceDocumentByIDEndpoint(library)
	libraryEndpoints.DeleteDocumentsByIDEndpoint = libraryClient.MakeDeleteDocumentsByIDEndpoint(library)

	ctx := context.Background()

	r := bone.New()
	r.SubRoute("/library", MakeLibraryHTTPHandler(ctx, libraryEndpoints))
	if err := http.ListenAndServe(*httpAddr, r); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
