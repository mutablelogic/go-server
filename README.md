# go-server

This repository implements a "monolithic generic task server", which can serve
requests over HTTP and FastCGI in addition to running "tasks" in the background.
Unlike many other servers, this one is composed of many "plugins" which can be
embedded within the server at compile time, or dynamically loaded as plugins
at runtime.

Standard plugins provided include:

  * __httpserver__ which provides a simple HTTP server and
    routing of requests to plugins;
  * __log__ which provides logging for requests and any other
    tasks;

The motivation for this module is to provide a generic server which
can be developed and scaled over time. Ultimately the running process
is a large "monolith" server which can be composed of many smaller 
"plugins", which can be connected together loosely (using a queue in between)
or tightly (by calling plugin methods directly).

## Requirements and Building

Any modern `go` compiler should be able to build the `server` command,
1.17 and above. It has been tested on MacOS and Linux. To build the server
and plugins, run:

```bash
[bash] git clone git@github.com:mutablelogic/go-server.git
[bash] cd go-server && make
```

This places all the binaries in the `build` directory. The folder structure
of the repository is as follows:

  * `cmd/server` contains the command line server tool. In order to build it,
    run `make server`. This places the binary in the `build` folder;
  * `etc` contains files which are used by the server, including a sample
    configuration file;
  * `pkg` contains the main code for the server and plugins;
  * `plugin` contains plugin bindings. To build the `httpserver` plugin for
    example run `make plugin/httpserver`. This places the plugin (with `.plugin` 
    file extension) in the `build` folder.

## Running the Server

You can run the server:

  1. With a HTTP server over network: You can specify TSL key and certificate
    to serve requests over a secure connection;
  2. With a HTTP server with FastCGI over a unix socket: You would want to do
    this if the server is behind a reverse proxy such as nginx.
  3. In a docker container, and expose the port outside the container. The docker
     container targets `amd64`, `arm64` and `arm` architectures on Linux.

## Project Status

This module is currently __in development__ and is not yet ready for any production environment.

## Community & License

  * [File an issue or question](http://github.com/mutablelogic/go-server/issues) on github.
  * Licensed under Apache 2.0, please read that license about using and forking. The main conditions require preservation of copyright and license notices. Contributors provide an express grant of patent rights. Licensed works, modifications, and larger works may be distributed under different terms and without source code.