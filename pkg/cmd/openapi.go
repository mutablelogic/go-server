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
	"net/url"

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

	u, err := url.JoinPath(endpoint, "openapi.html")
	if err != nil {
		return err
	}
	return browser.OpenURL(u)
}
