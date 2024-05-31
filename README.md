# go-server

This repository implements a "monolithic generic task server", which can serve
requests over HTTP and FastCGI. There is a "plugin architecture" which can be
embedded within the server at compile time, or dynamically loaded as plugins
at runtime.

Standard plugins provided include:

- [__httpserver__](pkg/httpserver) which provides a simple HTTP server and
  routing of requests to plugins;
- [__router__](pkg/handler/router) to route requests to different handlers;
- [__nginx__](pkg/handler/nginx) to manage a running nginx reverse proxy
  instance;
- [__static__](pkg/handler/static/) to serve static files;
- [__certmanager__](pkg/handler/certmanager) to manage trust and certificates.

The motivation for this module is to provide a generic server which
can be developed and scaled over time. Ultimately the running process
is a large "monolith" server which can be composed of many smaller
"plugins", which can be connected together loosely.

## Requirements and Building

Any modern `go` compiler should be able to build the `server` command,
1.22 and above. It has been tested on MacOS and Linux. To build the server
and plugins, run:

```bash
git clone git@github.com:mutablelogic/go-server.git
cd go-server && make
```

This places all the binaries in the `build` directory. There are several
other make targets:

- `make clean` to remove all build artifacts;
- `make test` to run all tests;
- `ARCH=<arm64|amd64> OS=<linux|darwin> make` to cross-compile the binary;
- `DOCKER_REPOSITORY=docker.io/user make docker` to build a docker image.
- `DOCKER_REPOSITORY=docker.io/user make docker-push` to push a docker image.

## Running the Server

You can run the server:

  1. With a HTTP server over network: You can specify TSL key and certificate
    to serve requests over a secure connection;
  2. With a HTTP server with FastCGI over a unix socket: You would want to do
    this if the server is behind a reverse proxy such as nginx.
  3. In a docker container, and expose the port outside the container. The docker
     container targets `amd64` and `arm64` architectures on Linux.

## Project Status

This module is currently __in development__ and is not yet ready for any production 
environment.

## Community & License

[File an issue or question](http://github.com/mutablelogic/go-server/issues) on github.

Licensed under Apache 2.0, please read that license about using and forking. The
main conditions require preservation of copyright and license notices. Contributors 
provide an express grant of patent rights. Licensed works, modifications, and larger 
works may be distributed under different terms and without source code.
