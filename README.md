# go-server

This module provides a generic server, which serves requests
over HTTP and FastCGI and can also run tasks in the background. Unlike
many other servers, this one is composed of many
"plugins" which can be added to the server.

Standard plugins provided include:

  * __httpserver__ which provides a simple HTTP server and
    routing of requests to plugins;
  * __log__ which provides logging for requests and any other
    tasks;
  * __env__ and __consul__ which allow retrieval of values from keys
    for use in configuration files;
  * __basicauth__ provides basic authentication for requests;
  * __ldapauth__ provides LDAP authentication for requests;
  * __static__ provides static file serving;
  * __renderer__ converts files into HTML documents;
  * __template__ provides dynamic file serving of documents through templates.

Many of these modules also provide a REST API for accessing information
and control, and there are a number of "front ends" developed for display
of plugin information in a web browser.

The motivation for this module is to provide a generic server which
can be developed and scaled over time. Ultimately the running process
is a large "monolith" server which can be composed of many smaller 
"plugins", which can be connected together loosely (using a queue in between)
or tightly (by calling plugin methods directly).

Maybe this design is a good balance between microservices and large (non-plugin) 
monoliths?

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
  * `npm` contains frontend NPM packages which can be built. To build the
    `mdns` frontend for example, run `make npm/mdns`. The compiled code
     is then in the `dist` folder of each NPM package. These may eventually be
    moved to a separate repository;
  * `pkg` contains the main code for the server and plugins;
  * `plugin` contains code for the plugins. To build the `httpserver` plugin for
    example run `make plugin/httpserver`. This places the plugin (with `.plugin` 
    file extension) in the `build` folder.

The `provider.go` file contains the interfaces required if you develop plugins.
More information about developing plugins is described below.

## Running the Server

You can run the server:

  1. With a HTTP server over network: You can specify TSL key and certificate
    to serve requests over a secure connection;
  2. With a HTTP server with FastCGI over a unix socket: You would want to do
    this if the server is behind a reverse proxy such as nginx.

It is most likely that in a production environment you would want to install the
server and any plugins with some sort of packaging (RPM or DEB) or from a Docker
container. More information about packaging can be found in the next section.

The `-help` argument provides more information on the command line options:

```bash
[bash]  /opt/go-server/bin/server -help

server: Monolith server

Usage:
  server <flags> config.yaml
  server -help
  server -help <plugin>

Flags:
  -addr string
    	Override path to unix socket or listening address

Version:
  URL: https://github.com/mutablelogic/go-server
  Version: v1.0.6
  Build Time: 2021-09-31T12:00:00Z
  Go: go1.17 (darwin/amd64)
```

## Configuration and Packaging

Packaging for binaries is available in the [pkg-server](https://github.com/mutablelogic/pkg-server)
repository for download.

Whilst you can run the server without a reverse proxy, it is recommended that
you use `nginx` or similar to serve the frontend files and communicate with the
server using FastCGI with a unix socket.

## Developing Plugins

You can develop different classes of plugins for the server, including:

  * A task plugin, which can runs in the background;
  * A frontend REST API plugin, which responds to requests over HTTP;
  * A renderer plugin, which can convert files into structured documents
    for display in a web browser or index into a search engine;
  * A frontend plugin, which serves embedded files statically;
  * A middleware plugin, which intercepts HTTP requests and can modify
    the request or response;
  * A logging plugin, which can log information from the server;
  * A keystore plugin, which can store and retrieve keys and other information.

Any plugin can utilize other plugins, or these plugin classes can be combined
into a single plugin.

TODO

## Project Status

This module is currently __in development__ and is not yet ready for any production environment.

## Community & License

  * [File an issue or question](http://github.com/mutablelogic/go-server/issues) on github.
  * Licensed under Apache 2.0, please read that license about using and forking. The main conditions require preservation of copyright and license notices. Contributors provide an express grant of patent rights. Licensed works, modifications, and larger works may be distributed under different terms and without source code.
