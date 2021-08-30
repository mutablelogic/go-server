# go-server

This module provides a generic server, which serves requests
over HTTP and FastCGI and can also run tasks in the background. Unlike
many other go servers, this one can be composed of many
"plugins" which can be added, developed and removed to the
server.

Standard plugins provided include:

  * __httpserver__ which provides a simple HTTP server and
    routing of requests to plugins;
  * __log__ which provides logging for requests and any other
    tasks;
  * __basicauth__ provides basic authentication for requests;
  * __ldapauth__ provides LDAP authentication for requests;
  * __static__ provides static file serving;
  * __template__ provides dynamic file serving through templates;
  * __sqlite__ provides SQLite database support;
  * __eventqueue__ provides a queue for messaging between plugins;
  * __mdns__ provides service discovery via mDNS.

Many of these modules also provide a REST API for accessing information
and controlling the plugin, and there are a number of "front ends"
developed for display in a web browser.

The motivation for this module is to provide a generic server which
can be developed and scaled over time. Ultimately the running process
is a large "monolith" server which can be composed of many smaller 
"plugins". Maybe this is a good balance between microservices and
large (non-plugin) monoliths.

## Requirements and Building

Any modern `go` compiler should be able to build the `server` command,
1.16 and above. It has been tested on MacOS and Linux.

In order to compile the front ends, `npm` is required which pulls in
additional dependencies.

TODO

## Running the Server

You can run the server:

  1. With a HTTP server over network: You can specify TSL key and certificate
    to serve requests over a secure connection;
  2. With a HTTP server with FastCGI over a unix socket: You would want to do
    this if the server is behind a reverse proxy such as nginx.

It is most likely that in a production environment you would want to install the
server and any plugins with some sort of packaging (RPM or DEB) or from a Docker
container. More information about packaging can be found in the next section.

TODO

## Configuration and Packaging

TODO

## Developing Plugins

TODO

## Project Status

This module is currently __in development__ and is not yet ready for any production environment.

## Community & License

  * [File an issue or question](http://github.com/djthorpe/go-server/issues) on github.
  * Licensed under Apache 2.0, please read that license about using and forking. The main conditions require preservation of copyright and license notices. Contributors provide an express grant of patent rights. Licensed works, modifications, and larger works may be distributed under different terms and without source code.
