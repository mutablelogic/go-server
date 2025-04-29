# go-server

This is an opinionated HTTP server which is designed to be used as a base for building
API and web applications. It is designed to be extensible and modular, allowing you to add
new features and functionality as needed. By default the server includes the following
features:

* Authentication of users using JWT tokens and API keys
* Connection to PostgreSQL databases
* Task queues for running background jobs
* Ability to manage the PostgreSQL database roles, databases, schemas and connections
* Prometheus metrics support

The idea is that you can use this server as a base for your own applications, and add your own
features and functionality as needed. More documentation soon on how to do that.

## Running

You can run the server in a docker container or build it from source. To run the latest
released version as a docker container:

```bash
docker run ghcr.io/mutablelogic/go-server:latest
```

This will print out the help message.

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

- [Go](https://golang.org/doc/install/source) (1.23 or later, not required for docker builds)
- [Make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/get-docker/)
- [NPM](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

### Makefile targets

Binaries are placed in the `build` directory.

| Target | Description |
|--------|-------------|
| `make all` | Build all plugins and the server |
| `make cmd/server` | Build the server binary |
| `make plugins` | Build the server binary |
| `make docker` | Build the docker container |
| `make docker-push` | Push the docker container to remote repository, assuming logged in |
| `make docker-version` | Prints out the docker container tag |
| `make test` | Runs unit amd coverage tests |
| `make unit-test` | Runs unit tests |
| `VERBOSE=1 make unit-test` | Runs unit tests with verbose output |
| `make coverage-test` | Reports code coverage |
| `make tidy` | Runs go mod tidy |
| `make clean` | Removes binaries and distribution files |

You can also affect the build by setting the following environment variables. For example,

```bash
GOOS=linux GOARCH=amd64 make
```

| Variable | Description |
|----------|-------------|
| `GOOS` | The target operating system for the build |
| `GOARCH` | The target architecture for the build |
| `BUILD_DIR` | The target architecture for the build |
| `VERBOSE` | Setting this flag will provide verbose output for unit tests |
| `VERSION` | Explicitly set the version |
| `DOCKER_REPO` | The docker repository to push to. Defaults to `ghcr.io/mutablelogic/go-server` |

