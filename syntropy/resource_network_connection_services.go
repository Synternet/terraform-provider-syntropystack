package syntropy

import (
	"context"
	"fmt"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = networkConnectionServiceResourceType{}
var _ tfsdk.Resource = networkConnectionServiceResource{}
var _ tfsdk.ResourceWithImportState = networkConnectionServiceResource{}

type networkConnectionServiceResourceType struct{}

type networkConnectionServiceResource struct {
	provider provider
}

func (t networkConnectionServiceResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Enables services inside connection group",
		Attributes: map[string]tfsdk.Attribute{
			"connection_group_id": {
				Description: "Unique identifier for the connection",
				Type:        types.Int64Type,
				Required:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"services": {
				Description: "List of network connection services to enable",
				Required:    true,
				Validators: []tfsdk.AttributeValidator{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Type:        types.Int64Type,
						Required:    true,
						Description: "Network connection service ID",
					},
					"enabled": {
						Type:        types.BoolType,
						Required:    true,
						Description: "Should network connection service be enabled?",
					},
				}),
			},
		},
	}, nil
}

func (t networkConnectionServiceResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return networkConnectionServiceResource{
		provider: provider,
	}, diags
}

func (r networkConnectionServiceResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan ConnectionService
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var changes []syntropy.AgentServicesUpdateChanges

	for _, service := range plan.Services {
		changes = append(changes, syntropy.AgentServicesUpdateChanges{
			AgentServiceSubnetId: int32(service.ID),
			IsEnabled:            service.Enabled,
		})
	}

	groupId := int32(plan.ConnectionGroupID.Value)
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesUpdate(ctx).V1NetworkConnectionsServicesUpdateRequest(syntropy.V1NetworkConnectionsServicesUpdateRequest{
		AgentConnectionGroupId: &groupId,
		Changes:                changes,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating connection service", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionServiceResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state ConnectionService
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesGet(ctx).Filter(strconv.FormatInt(state.ConnectionGroupID.Value, 10)).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while getting network connection service", err.Error())
		return
	}

	if len(connection.Data) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("Connection not found by ID = %s", strconv.FormatInt(state.ConnectionGroupID.Value, 10)), "")
		return
	}

	for _, stateServices := range state.Services {
		found := false
		for _, remoteServices := range connection.Data[0].AgentConnectionSubnets {
			if int32(stateServices.ID) == remoteServices.AgentServiceSubnetId {
				stateServices.Enabled = remoteServices.AgentConnectionSubnetIsEnabled
				found = true
				break
			}
		}
		if !found {
			stateServices.Enabled = false
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r networkConnectionServiceResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan ConnectionService
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var changes []syntropy.AgentServicesUpdateChanges

	for _, service := range plan.Services {
		changes = append(changes, syntropy.AgentServicesUpdateChanges{
			AgentServiceSubnetId: int32(service.ID),
			IsEnabled:            service.Enabled,
		})
	}

	groupId := int32(plan.ConnectionGroupID.Value)
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesUpdate(ctx).V1NetworkConnectionsServicesUpdateRequest(syntropy.V1NetworkConnectionsServicesUpdateRequest{
		AgentConnectionGroupId: &groupId,
		Changes:                changes,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating network connection service", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionServiceResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ConnectionService
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	var changes []syntropy.AgentServicesUpdateChanges

	for _, service := range state.Services {
		changes = append(changes, syntropy.AgentServicesUpdateChanges{
			AgentServiceSubnetId: int32(service.ID),
			IsEnabled:            false,
		})
	}

	if resp.Diagnostics.HasError() {
		return
	}
	groupId := int32(state.ConnectionGroupID.Value)
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesUpdate(ctx).V1NetworkConnectionsServicesUpdateRequest(syntropy.V1NetworkConnectionsServicesUpdateRequest{
		AgentConnectionGroupId: &groupId,
		Changes:                changes,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating network connection service", err.Error())
		return
	}
}

func (r networkConnectionServiceResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
