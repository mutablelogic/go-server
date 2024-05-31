# httpserver

This package provides a simple HTTP server that can be used to serve HTTP, HTTPS and FCGI requests.
For example:

```go
package main

import (
    "github.com/mutablelogic/go-server/pkg/httpserver"
)

func main() {
    // Create the server
    config := httpserver.Config{}
    server, err := config.New(context.Background())

    // Run the server for five seconds
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Serving HTTP requests

To serve HTTP requests, create a new server with the default configuration. By default, the
default router is of type `*http.ServeMux`. For example, the following creates a server that
listens on port 8080:

```go
func main() {
    server, err := httpserver.Config{ Listen: ":8080" }.New(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // Set the router handers
    server.Router().(*http.ServeMux).HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // ....
}
```

You can use a random port by setting the listen parameter to `:` or `:0`. You can find out which
port the server is listening on by calling `server.Addr()`. The timeout for serving requests can
be set using the `Timeout` parameter.

## Serving HTTPS requests

To serve HTTPS requests, create a new server with the default configuration and set the `TLS`
parameter to the path for the key and certificate files. For example, 

```go
func main() {
    server, err := httpserver.Config{
        TLS: httpserver.TLSConfig{
            Key: "server.key",
            Cert: "server.crt",
        },
    }.New(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // ....
}

```

## Serving FCGI requests

To serve requests over FastCGI using a unix socker, create a new server with the path to the socket:

```go
func main() {
    server, err := httpserver.Config{
        Listen: "socket.fcgi",
        Group: "http",
    }.New(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // ....
}
```

Set the `Owner` and `Group` configuration parameters to the user and group permissions for the socket file.

## Using a custom router

To use a custom router, set the `Router` parameter to the router you want to use. For example, to use
the `gorilla/mux` router:

```go
package main

import (
    httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
    mux "github.com/gorilla/mux"
)

func main() {
    server, err := httpserver.Config{
        Router: mux.NewRouter(),
    }.New(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // Set the router handers
    server.Router().(*mux.Router).HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // ....
}
```
