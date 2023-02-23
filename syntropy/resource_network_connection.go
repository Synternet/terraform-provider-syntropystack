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
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = networkConnectionResourceType{}
var _ tfsdk.Resource = networkConnectionResource{}
var _ tfsdk.ResourceWithImportState = networkConnectionResource{}

type networkConnectionResourceType struct{}

type networkConnectionResource struct {
	provider provider
}

func (t networkConnectionResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Creates connection between two Syntropy Platform agents",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier for the connection",
				Type:        types.Int64Type,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"agent_peer": {
				Description: "List of agent IDs for network connection",
				Type: types.SetType{
					ElemType: types.Int64Type,
				},
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					setvalidator.SizeBetween(2, 2),
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"sdn_enabled": {
				Description: "Should SDN be enabled?",
				Type:        types.BoolType,
				Optional:    true,
			},
			"services": {
				Description: "List of services inside in network connection",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:        types.StringType,
						Computed:    true,
						Description: "Network connection service name",
					},
					"id": {
						Type:        types.Int64Type,
						Computed:    true,
						Description: "Network connection service ID",
					},
					"ip": {
						Type:        types.StringType,
						Computed:    true,
						Description: "Network connection service IP",
					},
					"type": {
						Type:        types.StringType,
						Computed:    true,
						Description: "Network connection service type (Kubernetes, Docker, etc.)",
					},
					"enabled": {
						Type:        types.BoolType,
						Computed:    true,
						Description: "Is network connection service enabled?",
					},
					"agent_id": {
						Type:        types.Int64Type,
						Computed:    true,
						Description: "Network connection agent ID that service is created",
					},
				}),
			},
		},
	}, nil
}

func (t networkConnectionResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return networkConnectionResource{
		provider: provider,
	}, diags
}

func (r networkConnectionResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan NetworkConnection
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsCreateP2P(ctx).V1NetworkConnectionsCreateP2PRequest(syntropy.V1NetworkConnectionsCreateP2PRequest{
		AgentPairs: []syntropy.V1NetworkConnectionsCreateP2PRequestAgentPairsInner{
			{
				Agent2Id: int32(plan.AgentIds[0]),
				Agent1Id: int32(plan.AgentIds[1]),
			},
		},
		SdnEnabled: &plan.SdnEnabled.Value,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while creating network connection", err.Error())
		return
	}

	if len(connection.Data) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("Something went wrong while creating connection between agent_1_id=%d, agent_2_id=%d", plan.AgentIds[0], plan.AgentIds[1]), "Import network connection to state if connection already exists")
		return
	}

	connectionDetails, err := getOneConnectionDetails(ctx, *r.provider.client.ConnectionsApi, *connection.Data[0].AgentConnectionGroupId)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %d services", *connection.Data[0].AgentConnectionGroupId), err.Error())
		return
	}
	plan.ID = types.Int64{Value: int64(*connection.Data[0].AgentConnectionGroupId)}
	plan.Services = connectionDetails.Services
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state NetworkConnection
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.getConnectionGroupByAgentIDs(ctx, int32(state.AgentIds[0]), int32(state.AgentIds[1]))
	if err != nil {
		if err == ErrConnectionNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error while reading network connection", err.Error())
		return
	}

	connectionDetails, err := getOneConnectionDetails(ctx, *r.provider.client.ConnectionsApi, connection.AgentConnectionGroupId)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %d services", connection.AgentConnectionGroupId), err.Error())
		return
	}

	state.Services = connectionDetails.Services
	state.SdnEnabled = types.Bool{Value: connection.AgentConnectionGroupSdnEnabled}
	diags = resp.State.Set(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r networkConnectionResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan NetworkConnection
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsUpdate(ctx).V1NetworkConnectionsUpdateRequest(syntropy.V1NetworkConnectionsUpdateRequest{
		Changes: []syntropy.V1ConnectionUpdateChange{
			{
				ConnectionGroupId: int32(plan.ID.Value),
				IsSdnEnabled:      plan.SdnEnabled.Value,
			},
		},
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating network connection", err.Error())
		return
	}

	connectionDetails, err := getOneConnectionDetails(ctx, *r.provider.client.ConnectionsApi, int32(plan.ID.Value))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %d services", plan.ID.Value), err.Error())
		return
	}

	plan.Services = connectionDetails.Services

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data NetworkConnection
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsRemove(ctx).V1NetworkConnectionsRemoveRequest(syntropy.V1NetworkConnectionsRemoveRequest{
		AgentConnectionGroupIds: []int32{int32(data.ID.Value)},
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while deleting network connection", err.Error())
		return
	}
}

func (r networkConnectionResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r networkConnectionResource) getConnectionGroupByAgentIDs(ctx context.Context, agentID1 int32, agentID2 int32) (*syntropy.V1Connection, error) {
	resp, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsGet(ctx).Execute()
	if err != nil {
		return nil, err
	}

	for _, group := range resp.Data {
		if (group.Agent1.AgentId == agentID1 && group.Agent2.AgentId == agentID2) || (group.Agent1.AgentId == agentID2 && group.Agent2.AgentId == agentID1) {
			return &group, nil
		}
	}
	return nil, ErrConnectionNotFound
}
