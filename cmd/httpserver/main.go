package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"syscall"

	// Packages
	ctx "github.com/mutablelogic/go-server/pkg/context"
	static "github.com/mutablelogic/go-server/pkg/handler/static"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
)

var (
	port = flag.Int("port", 0, "Port to listen on")
	path = flag.String("path", "", "Path to serve")
)

func main() {
	flag.Parse()

	// Create a static file handler
	filesys, err := filesys()
	if err != nil {
		log.Fatal(err)
	}
	static, err := static.Config{FS: filesys, Dir: true}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Set the listen address
	listen := ":"
	if *port != 0 {
		listen = fmt.Sprintf(":%d", *port)
	}

	// Create the http server
	server, err := httpserver.Config{Listen: listen, Router: static.(http.Handler)}.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Run the server until we receive a signal
	ctx := ctx.ContextForSignal(os.Interrupt, syscall.SIGQUIT)
	log.Printf("Starting %q server on %q", server.Type(), server.Addr())
	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func filesys() (fs.FS, error) {
	if *path != "" {
		return os.DirFS(*path), nil
	} else if wd, err := os.Getwd(); err != nil {
		return nil, err
	} else {
		return os.DirFS(wd), nil
	}
}
