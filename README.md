# go-server

go-server is a Go module of packages for building HTTP servers and their command-line clients. Applications import it:

```bash
go get github.com/mutablelogic/go-server
```

## Packages

| Package | Description |
| ------- | ----------- |
| [`pkg/cmd`](pkg/cmd) | Command-line applications, with a server command, global flags and defaults which are kept between commands |
| [`pkg/httpserver`](pkg/httpserver) | The HTTP server, with its timeouts |
| [`pkg/httprouter`](pkg/httprouter) | The router, with cross-origin requests, and an OpenAPI specification of its paths |
| [`pkg/httprequest`](pkg/httprequest) | Reading requests: JSON, forms, queries, and files streamed from the request |
| [`pkg/httpresponse`](pkg/httpresponse) | Writing responses: JSON, errors, streams, and files, with their caching and preconditions |
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
| `make test` | Runs go vet, and the unit and coverage tests |
| `make vet` | Runs go vet |
| `make unit-test` | Runs the unit tests, verbosely with `VERBOSE=1` |
| `make race-test` | Runs the unit tests with the race detector |
| `make coverage-test` | Reports the tests' coverage |
| `make tidy` | Runs go mod tidy |
| `make clean` | Removes the build directory |

The tests run in GitHub Actions for pull requests and pushes to `main`, on Linux and Windows, and for releases on macOS too.
