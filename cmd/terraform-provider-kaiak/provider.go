package main

import (
	"context"
	"os"

	// Packages
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	httpclient "github.com/mutablelogic/go-server/pkg/provider/httpclient"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// kaiakProvider implements the Terraform provider for a running Kaiak server.
type kaiakProvider struct{}

// kaiakProviderModel maps provider schema data to a Go type.
type kaiakProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

var _ provider.Provider = (*kaiakProvider)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns a new provider instance. It is called by the plugin framework.
func New() provider.Provider {
	return &kaiakProvider{}
}

// resolveEndpoint returns the API endpoint from the environment, falling
// back to a sensible default.
func resolveEndpoint() string {
	if v := os.Getenv("KAIAK_ENDPOINT"); v != "" {
		return v
	}
	return "http://localhost:8084/api"
}

///////////////////////////////////////////////////////////////////////////////
// PROVIDER INTERFACE

func (p *kaiakProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kaiak"
}

func (p *kaiakProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Manage resources on a running Kaiak server.",
		Attributes: map[string]tfschema.Attribute{
			"endpoint": tfschema.StringAttribute{
				Description: "Base URL of the Kaiak server API (e.g. http://localhost:8084/api). " +
					"Can also be set via the KAIAK_ENDPOINT environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *kaiakProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config kaiakProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve endpoint: config value > environment variable > default
	endpoint := config.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = resolveEndpoint()
	}

	// Create the HTTP client
	client, err := httpclient.New(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Kaiak client", err.Error())
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources discovers resource types from the running Kaiak server and
// returns a factory for each one. The server must be reachable via
// KAIAK_ENDPOINT (or the default http://localhost:8084/api) at schema-
// discovery time (i.e. during terraform plan / apply).
func (p *kaiakProvider) Resources(_ context.Context) []func() resource.Resource {
	client, err := httpclient.New(resolveEndpoint())
	if err != nil {
		return nil
	}

	result, err := client.ListResources(context.Background(), schema.ListResourcesRequest{})
	if err != nil {
		return nil
	}

	factories := make([]func() resource.Resource, 0, len(result.Resources))
	for _, r := range result.Resources {
		meta := r // capture
		factories = append(factories, func() resource.Resource {
			return newDynamicResource(meta)
		})
	}
	return factories
}

func (p *kaiakProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewResourcesDataSource,
	}
}
