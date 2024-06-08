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
- [__auth__](pkg/handler/auth) to manage authentication and authorisation;
- [__tokenjar__](pkg/handler/tokenjar) to manage persistence of authorisation 
  tokens on disk;
- [__certmanager__](pkg/handler/certmanager) to manage trust and certificates.

The motivation for this module is to provide a generic server which
can be developed and scaled over time. Ultimately the running process
is a large "monolith" server which can be composed of many smaller
"plugins", which can be connected together loosely.

## Running the server

The easiest way to run an nginx reverse proxy server, with an API to
manage nginx configuration, is through docker:

```bash
docker run -p 8080:80 -v /var/lib/go-server:/data ghcr.io/mutablelogic/go-server
```

This will start a server on port 8080 and use `/var/lib/go-server` for persistent
data. Use API commands to manage the nginx configuration. Ultimately you'll 
want to develop your own plugins and can use this image as the base image for your 
own server.

When you first run the server, a "root" API token is created which is used to
authenticate API requests. You can find this token in the log output or by running
the following command:

```bash
docker exec <container-id> cat /data/tokenauth.json
```

## Requirements and Building

Any modern `go` compiler should be able to build the `server` command,
1.21 and above. It has been tested on MacOS and Linux. To build the server
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

## Project Status

This module is currently __in development__ and is not yet ready for any production 
environment.

## Community & License

[File an issue or question](http://github.com/mutablelogic/go-server/issues) on github.

Licensed under Apache 2.0, please read that license about using and forking. The
main conditions require preservation of copyright and license notices. Contributors 
provide an express grant of patent rights. Licensed works, modifications, and larger 
works may be distributed under different terms and without source code.
