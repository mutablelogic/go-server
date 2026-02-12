package main

import (
	"context"
	"fmt"

	// Packages
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	tfsdk "github.com/hashicorp/terraform-plugin-framework/tfsdk"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	httpclient "github.com/mutablelogic/go-server/pkg/provider/httpclient"
	schema "github.com/mutablelogic/go-server/pkg/provider/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// dynamicResource implements a Terraform resource whose schema is discovered
// at runtime from the Kaiak server.
type dynamicResource struct {
	client *httpclient.Client
	meta   schema.ResourceMeta
	infos  []attrInfo
}

// attrGetter is satisfied by tfsdk.Config, tfsdk.Plan, and tfsdk.State.
type attrGetter interface {
	GetAttribute(context.Context, path.Path, any) diag.Diagnostics
}

var _ resource.Resource = (*dynamicResource)(nil)
var _ resource.ResourceWithImportState = (*dynamicResource)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newDynamicResource(meta schema.ResourceMeta) *dynamicResource {
	infos := make([]attrInfo, 0, len(meta.Attributes))
	for _, a := range meta.Attributes {
		infos = append(infos, newAttrInfo(a))
	}
	return &dynamicResource{meta: meta, infos: infos}
}

// fullName returns the fully-qualified kaiak instance name.
func (r *dynamicResource) fullName(label string) string {
	return r.meta.Name + "." + label
}

///////////////////////////////////////////////////////////////////////////////
// RESOURCE INTERFACE

func (r *dynamicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.meta.Name
}

func (r *dynamicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s, _ := buildResourceSchema(r.meta.Name, r.meta.Attributes)
	resp.Schema = s
}

func (r *dynamicResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*httpclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *httpclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

///////////////////////////////////////////////////////////////////////////////
// CRUD

func (r *dynamicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read the instance label
	var name types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fullName := r.fullName(name.ValueString())

	// Create the instance on the server
	_, err := r.client.CreateResourceInstance(ctx, schema.CreateResourceInstanceRequest{
		Name: fullName,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create resource instance", err.Error())
		return
	}

	// Extract desired attributes from the plan and apply them
	attrs := r.extractAttrs(ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		_, _ = r.client.DestroyResourceInstance(ctx, fullName, false)
		return
	}

	if len(attrs) > 0 {
		_, err := r.client.UpdateResourceInstance(ctx, fullName, schema.UpdateResourceInstanceRequest{
			Attributes: attrs,
			Apply:      true,
		})
		if err != nil {
			_, _ = r.client.DestroyResourceInstance(ctx, fullName, false)
			resp.Diagnostics.AddError("Failed to apply attributes", err.Error())
			return
		}
	}

	// Read back the full state from the server
	r.writeState(ctx, fullName, name.ValueString(), &resp.State, &resp.Diagnostics)
}

func (r *dynamicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.writeState(ctx, id.ValueString(), name.ValueString(), &resp.State, &resp.Diagnostics)
}

func (r *dynamicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fullName := id.ValueString()

	// Extract desired attributes and apply them
	attrs := r.extractAttrs(ctx, req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateResourceInstance(ctx, fullName, schema.UpdateResourceInstanceRequest{
		Attributes: attrs,
		Apply:      true,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update resource instance", err.Error())
		return
	}

	r.writeState(ctx, fullName, name.ValueString(), &resp.State, &resp.Diagnostics)
}

func (r *dynamicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DestroyResourceInstance(ctx, id.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to destroy resource instance", err.Error())
	}
}

func (r *dynamicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by full qualified name (e.g. "httpstatic.docs")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE — extract terraform plan/config → kaiak State

// extractAttrs reads all non-readonly kaiak attributes from a terraform
// plan (or config). Block attributes are read by fetching the parent
// object first, then extracting individual fields.
func (r *dynamicResource) extractAttrs(ctx context.Context, src attrGetter, diags *diag.Diagnostics) schema.State {
	state := make(schema.State)

	// Top-level attributes
	for _, info := range r.infos {
		if info.attr.ReadOnly || info.tfBlock != "" {
			continue
		}
		switch info.attr.Type {
		case "bool":
			var v types.Bool
			diags.Append(src.GetAttribute(ctx, path.Root(info.tfField), &v)...)
			if !v.IsNull() && !v.IsUnknown() {
				state[info.kaiakName] = v.ValueBool()
			}
		case "int":
			var v types.Int64
			diags.Append(src.GetAttribute(ctx, path.Root(info.tfField), &v)...)
			if !v.IsNull() && !v.IsUnknown() {
				state[info.kaiakName] = v.ValueInt64()
			}
		default:
			var v types.String
			diags.Append(src.GetAttribute(ctx, path.Root(info.tfField), &v)...)
			if !v.IsNull() && !v.IsUnknown() {
				state[info.kaiakName] = v.ValueString()
			}
		}
	}

	// Block attributes — group by block name
	blockGroups := map[string][]attrInfo{}
	for _, info := range r.infos {
		if info.attr.ReadOnly || info.tfBlock == "" {
			continue
		}
		blockGroups[info.tfBlock] = append(blockGroups[info.tfBlock], info)
	}

	for blockName, infos := range blockGroups {
		var block types.Object
		diags.Append(src.GetAttribute(ctx, path.Root(blockName), &block)...)
		if block.IsNull() || block.IsUnknown() {
			continue
		}
		attrs := block.Attributes()
		for _, info := range infos {
			v, ok := attrs[info.tfField]
			if !ok {
				continue
			}
			switch info.attr.Type {
			case "bool":
				if bv, ok := v.(types.Bool); ok && !bv.IsNull() && !bv.IsUnknown() {
					state[info.kaiakName] = bv.ValueBool()
				}
			case "int":
				if iv, ok := v.(types.Int64); ok && !iv.IsNull() && !iv.IsUnknown() {
					state[info.kaiakName] = iv.ValueInt64()
				}
			default:
				if sv, ok := v.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
					state[info.kaiakName] = sv.ValueString()
				}
			}
		}
	}

	return state
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE — kaiak State → terraform state

// writeState fetches the instance from the server and populates
// the terraform state with the id, name, and all resource attributes.
func (r *dynamicResource) writeState(ctx context.Context, fullName, name string, tfState *tfsdk.State, diags *diag.Diagnostics) {
	result, err := r.client.GetResourceInstance(ctx, fullName)
	if err != nil {
		diags.AddError("Failed to read resource instance", err.Error())
		return
	}

	kaiakState := result.Instance.State

	// Fixed attributes
	diags.Append(tfState.SetAttribute(ctx, path.Root("id"), types.StringValue(fullName))...)
	diags.Append(tfState.SetAttribute(ctx, path.Root("name"), types.StringValue(name))...)

	// Top-level attributes
	for _, info := range r.infos {
		if info.tfBlock != "" {
			continue
		}
		v := kaiakState[info.kaiakName]
		diags.Append(tfState.SetAttribute(ctx, path.Root(info.tfField), kaiakValueToTF(v, info.attr.Type))...)
	}

	// Block attributes — set each block as a typed object
	blockGroups := map[string][]attrInfo{}
	for _, info := range r.infos {
		if info.tfBlock == "" {
			continue
		}
		blockGroups[info.tfBlock] = append(blockGroups[info.tfBlock], info)
	}

	for blockName, infos := range blockGroups {
		attrTypes := make(map[string]attr.Type, len(infos))
		attrValues := make(map[string]attr.Value, len(infos))
		hasValue := false

		for _, info := range infos {
			attrTypes[info.tfField] = kaiakTypeToAttrType(info.attr.Type)
			if v, ok := kaiakState[info.kaiakName]; ok && v != nil {
				hasValue = true
				attrValues[info.tfField] = kaiakValueToTF(v, info.attr.Type)
			} else {
				attrValues[info.tfField] = kaiakNullValue(info.attr.Type)
			}
		}

		if hasValue {
			obj, d := types.ObjectValue(attrTypes, attrValues)
			diags.Append(d...)
			diags.Append(tfState.SetAttribute(ctx, path.Root(blockName), obj)...)
		} else {
			diags.Append(tfState.SetAttribute(ctx, path.Root(blockName), types.ObjectNull(attrTypes))...)
		}
	}
}
