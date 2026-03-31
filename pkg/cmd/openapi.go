// Copyright 2026 David Thorpe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	// Packages
	server "github.com/mutablelogic/go-server"
	browser "github.com/pkg/browser"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type OpenAPICommands struct {
	OpenAPI OpenAPICommand `cmd:"" name:"openapi" help:"Show OpenAPI documentation in a browser, or output the spec as JSON or YAML." group:"SERVER"`
}

type OpenAPICommand struct {
	JSON bool `name:"json" xor:"format" help:"Output OpenAPI spec as JSON."`
	YAML bool `name:"yaml" xor:"format" help:"Output OpenAPI spec as YAML."`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *OpenAPICommand) Run(ctx server.Cmd) error {
	endpoint, _, err := ctx.ClientEndpoint()
	if err != nil {
		return err
	}

	// When --json or --yaml is set, fetch the spec and print it to stdout.
	// Otherwise open the HTML documentation in a browser.
	switch {
	case cmd.JSON:
		return fetchAndPrint(endpoint, "openapi.json")
	case cmd.YAML:
		return fetchAndPrint(endpoint, "openapi.yaml")
	default:
		u, err := url.JoinPath(endpoint, "openapi.html")
		if err != nil {
			return err
		}
		return browser.OpenURL(u)
	}
}

// fetchAndPrint fetches the given path relative to endpoint and copies the
// response body to stdout.
func fetchAndPrint(endpoint, path string) error {
	u, err := url.JoinPath(endpoint, path)
	if err != nil {
		return err
	}
	resp, err := http.Get(u) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", path, resp.Status)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}
