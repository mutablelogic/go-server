# nginx

Handler to start, stop, test and reload nginx programatically, assuming you have nginx installed.
The following code will start an nginx server, and then stop it after 10 seconds:

```go
package main

import (
    "github.com/mutablelogic/go-server/pkg/handler/nginx"
)

func main() { 
    if nginx, err := nginx.New(nginx.Config{}); err != nil {
        log.Fatal(err)
    }

    // Create a cancellable context
    ctx, cancel := context.WithTimeout(context.Background(),10 * time.Second)
    defer cancel()

    // Run nginx in foreground until timeout is reached
    if err := nginx.Run(ctx); err != nil {
        t.Error(err)
    }
}
```

Some additional methods can be used to stop, test and reload the server:

```go
type Nginx interface {
    // test the configuration and return an error if it fails
    Test() error

    // test the configuration and then reload it (the SIGHUP signal)
    Reload() error

    // reopen log files (the SIGUSR1 signal)
    Reopen() error

    // return the nginx version string
    Version() string
}
```

## Configuration

The configuration for the nginx server is embedded (no external configuration files are required). By default,
a **Hello, World** static server is setup on port 80. In reality, you'll want to override this configuration
with your own.

TODO

## API

The commands can be called through an API.

| Method | Path         | Scope | Description |
|--------|--------------|-------|-------------|
| GET    | /            | read  | Return the nginx version and uptime |
| PUT    | /test        | write | Test the server configuration |
| PUT    | /reload      | write | Test the configuration and then reload it|
| PUT    | /reopen      | write | Reopen log files |
| GET    | /config      | read  | Read the current set of configurations |
| GET    | /config/{id} | read  | Read a specific configuration file |
| DELETE | /config/{id} | write | Delete a configuration, and reload |
| POST   | /config/{id} | write | Create a new configuration, then reload |
| PATCH  | /config/{id} | write | Update a configuration enabled or body, and reload on change |

The body of the POST request should be a JSON object with the following fields:

- `enabled`: A boolean value to enable or disable the configuration
- `body`: A string value which contains the content of the configuration file

The scopes are (not yet implemented):

- read: `github.com/mutablelogic/go-server/handler/nginx.read`
- write: `github.com/mutablelogic/go-server/handler/nginx.write`
