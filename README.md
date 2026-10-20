# go-server

This is an opinionated HTTP server which is designed to be used as a base for building
API and web applications. It is designed to be extensible and modular, allowing you to add
new features and functionality as needed. By default the server includes the following
features:

* Anthentication of users using Cognito, Google and LDAP
* Authorization of users using JWT tokens and API keys
* Task queues for running background jobs
* Connection to PostgreSQL databases
* Ability to manage the PostgreSQL database roles, databases, schemas and connections
* Prometheus metrics support

The idea is that you can use this server as a base for your own applications, and add your own
features and functionality as needed. More documentation soon on how to do that.

## Using

go-server is a Go module, which applications import, rather than an application or a docker image:

```bash
go get github.com/mutablelogic/go-server
```

### Bootstrapping the database server and roles

More information about setting up the databses TODO

### Creating a new database

Information about setting up a new database

## Building

### Download and build

```bash
git clone github.com/mutablelogic/go-server
cd go-server
make
```

The plugins and the `server` binary will be built in the `build` directory.

### Build requirements

You need the following three tools installed to build the server:

* [Go](https://golang.org/doc/install/source), the version in `go.mod`
* [Make](https://www.gnu.org/software/make/)

### Makefile targets

Binaries are placed in the `build` directory.

| Target | Description |
|--------|-------------|
| `make all` | Build all plugins and the server |
| `make cmd/server` | Build the server binary |
| `make plugins` | Build the server binary |
| `make test` | Runs go vet, and unit and coverage tests |
| `make vet` | Runs go vet |
| `make unit-test` | Runs unit tests |
| `make race-test` | Runs unit tests with the race detector |
| `VERBOSE=1 make unit-test` | Runs unit tests with verbose output |
| `make coverage-test` | Reports code coverage |
| `make tidy` | Runs go mod tidy |
| `make clean` | Removes binaries and distribution files |

You can also affect the build by setting the following environment variables. For example,

```bash
OS=linux ARCH=amd64 make
```

| Variable | Description |
|----------|-------------|
| `OS` | The target operating system for the build |
| `ARCH` | The target architecture for the build |
| `BUILD_DIR` | The target architecture for the build |
| `VERBOSE` | Setting this flag will provide verbose output for unit tests |
| `VERSION` | Explicitly set the version |
