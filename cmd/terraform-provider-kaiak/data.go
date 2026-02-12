package main

import (
	"context"
	"fmt"

	// Packages
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	httpclient "github.com/mutablelogic/go-server/pkg/provider/httpclient"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// resourcesDataSource implements the kaiak_resources data source.
type resourcesDataSource struct {
	client *httpclient.Client
}

// resourcesDataSourceModel maps the data source schema to Go types.
type resourcesDataSourceModel struct {
	Type         types.String              `tfsdk:"type"`
	ProviderName types.String              `tfsdk:"provider_name"`
	Version      types.String              `tfsdk:"version"`
	Resources    []resourceDataSourceModel `tfsdk:"resources"`
}

// resourceDataSourceModel describes a single resource type in the list.
type resourceDataSourceModel struct {
	Name       types.String                  `tfsdk:"name"`
	Attributes []attributeDataSourceModel    `tfsdk:"attributes"`
	Instances  []instanceMetaDataSourceModel `tfsdk:"instances"`
}

// attributeDataSourceModel describes a resource attribute.
type attributeDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Required    types.Bool   `tfsdk:"required"`
	ReadOnly    types.Bool   `tfsdk:"readonly"`
	Sensitive   types.Bool   `tfsdk:"sensitive"`
	Reference   types.Bool   `tfsdk:"reference"`
}

// instanceMetaDataSourceModel describes an existing instance of a resource type.
type instanceMetaDataSourceModel struct {
	Name     types.String `tfsdk:"name"`
	Resource types.String `tfsdk:"resource"`
	ReadOnly types.Bool   `tfsdk:"readonly"`
}

var _ datasource.DataSource = (*resourcesDataSource)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewResourcesDataSource() datasource.DataSource {
	return &resourcesDataSource{}
}

///////////////////////////////////////////////////////////////////////////////
// DATA SOURCE INTERFACE

func (d *resourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resources"
}

func (d *resourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		Description: "Lists available resource types and their instances on a running Kaiak server.",
		Attributes: map[string]tfschema.Attribute{
			"type": tfschema.StringAttribute{
				Description: "Filter by resource type name (e.g. \"httpserver\"). Omit for all types.",
				Optional:    true,
			},
			"provider_name": tfschema.StringAttribute{
				Description: "Provider name (computed from the server).",
				Computed:    true,
			},
			"version": tfschema.StringAttribute{
				Description: "Provider version (computed from the server).",
				Computed:    true,
			},
			"resources": tfschema.ListNestedAttribute{
				Description: "The list of resource types.",
				Computed:    true,
				NestedObject: tfschema.NestedAttributeObject{
					Attributes: map[string]tfschema.Attribute{
						"name": tfschema.StringAttribute{
							Description: "Resource type name.",
							Computed:    true,
						},
						"attributes": tfschema.ListNestedAttribute{
							Description: "Schema attributes for this resource type.",
							Computed:    true,
							NestedObject: tfschema.NestedAttributeObject{
								Attributes: map[string]tfschema.Attribute{
									"name":        tfschema.StringAttribute{Computed: true},
									"type":        tfschema.StringAttribute{Computed: true},
									"description": tfschema.StringAttribute{Computed: true},
									"required":    tfschema.BoolAttribute{Computed: true},
									"readonly":    tfschema.BoolAttribute{Computed: true},
									"sensitive":   tfschema.BoolAttribute{Computed: true},
									"reference":   tfschema.BoolAttribute{Computed: true},
								},
							},
						},
						"instances": tfschema.ListNestedAttribute{
							Description: "Existing instances of this resource type.",
							Computed:    true,
							NestedObject: tfschema.NestedAttributeObject{
								Attributes: map[string]tfschema.Attribute{
									"name":     tfschema.StringAttribute{Computed: true},
									"resource": tfschema.StringAttribute{Computed: true},
									"readonly": tfschema.BoolAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *resourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*httpclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *httpclient.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *resourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resourcesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the list request
	listReq := schema.ListResourcesRequest{}
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		t := config.Type.ValueString()
		listReq.Type = &t
	}

	result, err := d.client.ListResources(ctx, listReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list resources", err.Error())
		return
	}

	// Map the response into the model
	config.ProviderName = types.StringValue(result.Provider)
	config.Version = types.StringValue(result.Version)
	config.Resources = make([]resourceDataSourceModel, 0, len(result.Resources))

	for _, r := range result.Resources {
		res := resourceDataSourceModel{
			Name:       types.StringValue(r.Name),
			Attributes: make([]attributeDataSourceModel, 0, len(r.Attributes)),
			Instances:  make([]instanceMetaDataSourceModel, 0, len(r.Instances)),
		}
		for _, a := range r.Attributes {
			res.Attributes = append(res.Attributes, attributeDataSourceModel{
				Name:        types.StringValue(a.Name),
				Type:        types.StringValue(a.Type),
				Description: types.StringValue(a.Description),
				Required:    types.BoolValue(a.Required),
				ReadOnly:    types.BoolValue(a.ReadOnly),
				Sensitive:   types.BoolValue(a.Sensitive),
				Reference:   types.BoolValue(a.Reference),
			})
		}
		for _, i := range r.Instances {
			res.Instances = append(res.Instances, instanceMetaDataSourceModel{
				Name:     types.StringValue(i.Name),
				Resource: types.StringValue(i.Resource),
				ReadOnly: types.BoolValue(i.ReadOnly),
			})
		}
		config.Resources = append(config.Resources, res)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
