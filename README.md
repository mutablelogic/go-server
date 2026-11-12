# go-server

go-server is a Go module of packages for building HTTP servers and their command-line clients. Applications import it:

```bash
go get github.com/mutablelogic/go-server
```

## Packages

| Package | Description |
| ------- | ----------- |
| [`broadcaster`](broadcaster) | Publishes changes, such as to the rows of a server's tables, and streams them to clients at `/changes`, with a client and a `changes` command |
| [`config`](config) | Serves the server's configuration at `/config.json`, for its frontend: its API's prefix, its name, its version, its documentation, and its public endpoint |
| [`frontend`](frontend) | The `@mutablelogic/go-server` TypeScript package, of the configuration and the broadcaster's changes, for the frontends of servers built on go-server |
| [`test`](test) | A HTTP test server of an API, which checks every request and response against its OpenAPI specification |
| [`pkg/cmd`](pkg/cmd) | Command-line applications, with a server command, global flags and defaults which are kept between commands |
| [`pkg/httpserver`](pkg/httpserver) | The HTTP server, with its timeouts |
| [`pkg/httprouter`](pkg/httprouter) | The router, with cross-origin requests, and an OpenAPI specification of its paths |
| [`pkg/httprequest`](pkg/httprequest) | Reading requests: JSON, forms, queries, and files streamed from the request |
| [`pkg/httpresponse`](pkg/httpresponse) | Writing responses: JSON, errors, streams with a deadline for each write, and files, with their caching and preconditions |
| [`pkg/jsonschema`](pkg/jsonschema) | JSON schemas of Go types, and validating JSON with them |
| [`pkg/openapi`](pkg/openapi) | OpenAPI specifications, and their documentation in markdown |
| [`pkg/otel`](pkg/otel) | OpenTelemetry tracing |
| [`pkg/logger`](pkg/logger) | Structured logging |
| [`pkg/tui`](pkg/tui) | Tables and other output for terminals |
| [`pkg/types`](pkg/types) | Common types, such as files, content types and headers |
| [`pkg/version`](pkg/version) | The version of a build |

## Building and testing

You need [Go](https://go.dev/doc/install), the version in `go.mod`, and [Make](https://www.gnu.org/software/make/).

| Target | Description |
| ------ | ----------- |
| `make` | Builds the commands in `cmd` into `build`, such as `build/server` |
| `make test` | Runs go vet, and the unit and coverage tests |
| `make vet` | Runs go vet |
| `make unit-test` | Runs the unit tests, with `TEST_FLAGS`, such as `TEST_FLAGS=-v` |
| `make race-test` | Runs the unit tests with the race detector |
| `make coverage-test` | Reports the tests' coverage, in `build/coverage.txt` |
| `make tidy` | Runs go mod tidy |
| `make frontend` | Builds the TypeScript package, with npm |
| `make frontend-test` | Tests the TypeScript package |
| `make clean` | Removes the build directory |

Go runs with `GOWORK=off`, so a Go workspace doesn't change what's built. Set `GOWORK` to use one.

`build/server` runs a server with go-server's APIs, such as the broadcaster's stream of changes, with `server run`, and is their client, such as with `server changes`.

The tests run in GitHub Actions for pull requests and pushes to `main` on Linux, and for releases on Windows and macOS too.
