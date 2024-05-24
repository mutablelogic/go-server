package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"syscall"

	// Packages
	ctx "github.com/mutablelogic/go-server/pkg/context"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
)

func main() {
	server, err := httpserver.Config{Listen: ":"}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Set the router handers
	server.Router().(*http.ServeMux).HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Run the server until we receive a signal
	ctx := ctx.ContextForSignal(os.Interrupt, syscall.SIGQUIT)
	log.Printf("Starting %q server on %q", server.Type(), server.Addr())
	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
