# static

Static File Handler from a filesystem. For example:

```go
package main

import (
    "github.com/mutablelogic/go-server/pkg/httpserver"
)

func main() {
    static, err := static.Config{FS: os.DirFS("/"), DirListing: true, Prefix: ""}.New(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    // ...
}
```

Passing `DirListing` as `true` will serve a basic directory listing for a filesystem. Otherwise,
the handler will serve files from the directory `FS` only, and serve the `index.html`
file is lieu of a directory, if it exists.

The configuration option `Prefix` will root the handler at a specific subdirectory.
