package syntropy

import (
	"context"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = networkConnectionSubnetResourceType{}
var _ tfsdk.Resource = networkConnectionSubnetResource{}
var _ tfsdk.ResourceWithImportState = networkConnectionSubnetResource{}

type networkConnectionSubnetResourceType struct{}

type networkConnectionSubnetResource struct {
	provider provider
}

func (t networkConnectionSubnetResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Enables connection services inside connection group",
		Attributes: map[string]tfsdk.Attribute{
			"connection_group_id": {
				Description: "Connection group ID",
				Type:        types.Int64Type,
				Required:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"subnet_id": {
				Description: "Subnet ID",
				Type:        types.Int64Type,
				Required:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"enable": {
				Description: "Should service be enabled",
				Type:        types.BoolType,
				Required:    true,
			},
		},
	}, nil
}

func (t networkConnectionSubnetResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return networkConnectionSubnetResource{
		provider: provider,
	}, diags
}

func (r networkConnectionSubnetResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan NetworkConnectionSubnet
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := int32(plan.ConnectionGroupID.Value)
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesUpdate(ctx).V1NetworkConnectionsServicesUpdateRequest(syntropy.V1NetworkConnectionsServicesUpdateRequest{
		AgentConnectionGroupId: &groupId,
		Changes: []syntropy.AgentServicesUpdateChanges{
			{
				AgentServiceSubnetId: int32(plan.SubnetID.Value),
				IsEnabled:            plan.Enable.Value,
			},
		},
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating connection service", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionSubnetResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state NetworkConnectionSubnet
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	enabled, err := r.isServiceEnabled(ctx, strconv.FormatInt(state.ConnectionGroupID.Value, 10), state.SubnetID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting network connection service", err.Error())
		return
	}

	state.Enable = types.Bool{Value: enabled}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r networkConnectionSubnetResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan NetworkConnectionSubnet
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupId := int32(plan.ConnectionGroupID.Value)
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesUpdate(ctx).V1NetworkConnectionsServicesUpdateRequest(syntropy.V1NetworkConnectionsServicesUpdateRequest{
		AgentConnectionGroupId: &groupId,
		Changes: []syntropy.AgentServicesUpdateChanges{
			{
				AgentServiceSubnetId: int32(plan.SubnetID.Value),
				IsEnabled:            plan.Enable.Value,
			},
		},
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating network connection service", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionSubnetResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state NetworkConnectionSubnet
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	groupId := int32(state.ConnectionGroupID.Value)
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesUpdate(ctx).V1NetworkConnectionsServicesUpdateRequest(syntropy.V1NetworkConnectionsServicesUpdateRequest{
		AgentConnectionGroupId: &groupId,
		Changes: []syntropy.AgentServicesUpdateChanges{
			{
				AgentServiceSubnetId: int32(state.SubnetID.Value),
				IsEnabled:            false,
			},
		},
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating network connection service", err.Error())
		return
	}
}

func (r networkConnectionSubnetResource) isServiceEnabled(ctx context.Context, groupId string, subnetID int64) (bool, error) {
	connection, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsServicesGet(ctx).Filter(groupId).Execute()
	if err != nil {
		return false, err
	}

	enabled := false
	for _, subnet := range connection.Data[0].AgentConnectionSubnets {
		if int64(subnet.AgentServiceSubnetId) == subnetID {
			return subnet.AgentConnectionSubnetIsEnabled, nil
		}
	}
	return enabled, nil
}

func (r networkConnectionSubnetResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
