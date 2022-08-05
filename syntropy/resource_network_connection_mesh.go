package syntropy

import (
	"context"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = networkConnectionMeshResourceType{}
var _ tfsdk.Resource = networkConnectionMeshResource{}
var _ tfsdk.ResourceWithImportState = networkConnectionMeshResource{}

type networkConnectionMeshResourceType struct{}

type networkConnectionMeshResource struct {
	provider provider
}

func (t networkConnectionMeshResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Creates network mesh between agents",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Network connection mesh ID",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"agent_ids": {
				Description: "List of agent IDs for network connection mesh",
				Type: types.SetType{
					ElemType: types.NumberType,
				},
				Required: true,
			},
			"sdn_enabled": {
				Description: "Should SDN be enabled?",
				Type:        types.BoolType,
				Optional:    true,
			},
			"connections": {
				Description: "Created connections",
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"agent_1_id": {
						Type:        types.NumberType,
						Computed:    true,
						Description: "Agent 1 ID",
					},
					"agent_2_id": {
						Type:        types.NumberType,
						Computed:    true,
						Description: "Agent 2 ID",
					},
					"agent_connection_group_id": {
						Type:        types.NumberType,
						Computed:    true,
						Description: "Agent connection group ID",
					},
				}),
			},
		},
	}, nil
}

func (t networkConnectionMeshResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return networkConnectionMeshResource{
		provider: provider,
	}, diags
}

func (r networkConnectionMeshResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan NetworkConnectionMeshData
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentList []syntropy.V1NetworkConnectionsCreateMeshRequestAgentIds
	for _, i := range plan.AgentIds {
		agentList = append(agentList, syntropy.V1NetworkConnectionsCreateMeshRequestAgentIds{
			AgentId: i,
		})
	}

	connections, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsCreateMesh(ctx).V1NetworkConnectionsCreateMeshRequest(syntropy.V1NetworkConnectionsCreateMeshRequest{
		AgentIds:   agentList,
		SdnEnabled: &plan.SdnEnabled.Value,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while creating network connection mesh", err.Error())
		return
	}

	var connResults []Connection
	for _, conn := range connections.Data {
		connResults = append(connResults, Connection{
			Agent1ID:          *conn.Agent1Id,
			Agent2ID:          *conn.Agent2Id,
			ConnectionGroupID: *conn.AgentConnectionGroupId,
		})
	}

	plan.ID = types.String{Value: uuid.New().String()}
	plan.Connections = connResults

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionMeshResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state NetworkConnectionMeshData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r networkConnectionMeshResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	resp.Diagnostics.AddError("Not implemented yet", "")
	return

	var plan NetworkConnectionMeshData

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.AddError("Not implemented yet", "")
	if resp.Diagnostics.HasError() {
		return
	}

	var agentList []syntropy.V1NetworkConnectionsCreateMeshRequestAgentIds
	for _, i := range plan.AgentIds {
		agentList = append(agentList, syntropy.V1NetworkConnectionsCreateMeshRequestAgentIds{
			AgentId: i,
		})
	}

	newConnections, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsCreateMesh(ctx).V1NetworkConnectionsCreateMeshRequest(syntropy.V1NetworkConnectionsCreateMeshRequest{
		AgentIds:   agentList,
		SdnEnabled: &plan.SdnEnabled.Value,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating network connection mesh", err.Error())
		return
	}

	var connResults []Connection
	for _, conn := range newConnections.Data {
		connResults = append(connResults, Connection{
			Agent1ID:          *conn.Agent1Id,
			Agent2ID:          *conn.Agent2Id,
			ConnectionGroupID: *conn.AgentConnectionGroupId,
		})
	}

	plan.Connections = append(plan.Connections, connResults...)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionMeshResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	resp.Diagnostics.AddError("Not implemented yet", "")
	return

	var data NetworkConnectionMeshData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r networkConnectionMeshResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
