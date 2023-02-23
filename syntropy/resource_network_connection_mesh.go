package syntropy

import (
	"context"
	"fmt"
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
				Description: "Network connection mesh ID randomly generated",
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
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Required: true,
			},
			"sdn_enabled": {
				Description: "Should SDN be enabled?",
				Type:        types.BoolType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Optional: true,
			},
			"connections": {
				Description: "List of network connections created by mesh resource",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"agent_1_id": {
						Type:        types.Int64Type,
						Computed:    true,
						Description: "Agent 1 ID",
					},
					"agent_2_id": {
						Type:        types.Int64Type,
						Computed:    true,
						Description: "Agent 2 ID",
					},
					"connection_group_id": {
						Type:        types.Int64Type,
						Computed:    true,
						Description: "Unique identifier for the connection",
					},
					"services": {
						Description: "List of services inside in network connection mesh",
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
	var plan NetworkConnectionMesh
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentList []syntropy.V1NetworkConnectionsCreateMeshRequestAgentIdsInner
	for _, i := range plan.AgentIds {
		agentList = append(agentList, syntropy.V1NetworkConnectionsCreateMeshRequestAgentIdsInner{
			AgentId: i,
		})
	}

	_, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsCreateMesh(ctx).V1NetworkConnectionsCreateMeshRequest(syntropy.V1NetworkConnectionsCreateMeshRequest{
		AgentIds:   agentList,
		SdnEnabled: &plan.SdnEnabled.Value,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while creating network mesh", err.Error())
		return
	}

	connections, err := r.GetConnectionsListByAgentID(ctx, plan.AgentIds)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting network mesh connections", err.Error())
		return
	}

	var connectionIDs []int32
	for _, conn := range connections {
		connectionIDs = append(connectionIDs, conn.ConnectionGroupID)
	}

	connectionDetails, err := getMultipleConnectionDetails(ctx, *r.provider.client.ConnectionsApi, connectionIDs)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %v services", connectionIDs), err.Error())
		return
	}

	// Apply services to existing connections array. We update existing array to avoid creating array copy
	for i := range connections {
		for j := range connectionDetails {
			if connections[i].ConnectionGroupID == connectionDetails[j].ConnectionGroupID {
				connections[i].Services = connectionDetails[j].Services
				break
			}
		}
	}

	plan.ID = types.String{Value: uuid.New().String()}
	plan.Connections = connections

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionMeshResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state NetworkConnectionMesh
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connections, err := r.GetConnectionsListByAgentID(ctx, state.AgentIds)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting network mesh connections", err.Error())
		return
	}

	expectedConnections := sumOfNaturalNumbers(len(state.AgentIds))
	// If expected connections count not equal to returned connections count this means that changes were made outside
	// terraform. In this case we need to force terraform to re-run apply
	if expectedConnections != len(connections) {
		resp.State.RemoveResource(ctx)
		return
	}

	var connectionIDs []int32
	for _, conn := range connections {
		connectionIDs = append(connectionIDs, conn.ConnectionGroupID)
	}

	connectionDetails, err := getMultipleConnectionDetails(ctx, *r.provider.client.ConnectionsApi, connectionIDs)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %v services", connectionIDs), err.Error())
		return
	}

	// Apply services to existing connections array. We update existing array to avoid creating array copy
	for i := range connections {
		for j := range connectionDetails {
			if connections[i].ConnectionGroupID == connectionDetails[j].ConnectionGroupID {
				connections[i].Services = connectionDetails[j].Services
				break
			}
		}
	}

	state.Connections = connections

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r networkConnectionMeshResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state NetworkConnectionMeshEdit
	ctx = r.provider.createAuthContext(ctx)

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.FindAndDeleteOldConnections(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var agentList []syntropy.V1NetworkConnectionsCreateMeshRequestAgentIdsInner
	for _, i := range plan.AgentIds {
		agentList = append(agentList, syntropy.V1NetworkConnectionsCreateMeshRequestAgentIdsInner{
			AgentId: i,
		})
	}

	_, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsCreateMesh(ctx).V1NetworkConnectionsCreateMeshRequest(syntropy.V1NetworkConnectionsCreateMeshRequest{
		AgentIds:   agentList,
		SdnEnabled: &plan.SdnEnabled.Value,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while creating network mesh", err.Error())
		return
	}

	connections, err := r.GetConnectionsListByAgentID(ctx, plan.AgentIds)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting network mesh connections", err.Error())
		return
	}

	var connectionIDs []int32
	for _, conn := range connections {
		connectionIDs = append(connectionIDs, conn.ConnectionGroupID)
	}

	connectionDetails, err := getMultipleConnectionDetails(ctx, *r.provider.client.ConnectionsApi, connectionIDs)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %v services", connectionIDs), err.Error())
		return
	}

	// Apply services to existing connections array. We update existing array to avoid creating array copy
	for i := range connections {
		for j := range connectionDetails {
			if connections[i].ConnectionGroupID == connectionDetails[j].ConnectionGroupID {
				connections[i].Services = connectionDetails[j].Services
				break
			}
		}
	}

	newState := NetworkConnectionMesh{
		ID:          plan.ID,
		AgentIds:    plan.AgentIds,
		Connections: connections,
		SdnEnabled:  plan.SdnEnabled,
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r networkConnectionMeshResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data NetworkConnectionMesh
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	deleteReq := syntropy.V1NetworkConnectionsRemoveRequest{}
	for _, a := range data.Connections {
		deleteReq.AgentConnectionGroupIds = append(deleteReq.AgentConnectionGroupIds, int32(a.ConnectionGroupID))
	}

	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsRemove(ctx).V1NetworkConnectionsRemoveRequest(deleteReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while deleting network mesh connections", err.Error())
		return
	}
}

func (r networkConnectionMeshResource) FindAndDeleteOldConnections(ctx context.Context, state NetworkConnectionMeshEdit, plan NetworkConnectionMeshEdit) diag.Diagnostics {
	var (
		diags           = diag.Diagnostics{}
		deleteRequest   = syntropy.V1NetworkConnectionsRemoveRequest{}
		connectionsList []Connection
	)

	diags = state.Connections.ElementsAs(ctx, &connectionsList, false)
	if diags.HasError() {
		return diags
	}

	for i := 0; i < len(state.AgentIds); i++ {
		found := false
		for j := 0; j < len(plan.AgentIds); j++ {
			if state.AgentIds[i] == plan.AgentIds[j] {
				found = true
				break
			}
		}
		if !found {
			for _, conn := range connectionsList {
				if state.AgentIds[i] == conn.Agent1ID || state.AgentIds[i] == conn.Agent2ID {
					deleteRequest.AgentConnectionGroupIds = append(deleteRequest.AgentConnectionGroupIds, int32(conn.ConnectionGroupID))
				}
			}
		}
	}

	if len(deleteRequest.AgentConnectionGroupIds) == 0 {
		return nil
	}

	_, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsRemove(ctx).V1NetworkConnectionsRemoveRequest(deleteRequest).Execute()
	if err != nil {
		diags.AddError("Error while deleting network mesh connections", err.Error())
		return diags
	}

	return nil
}

func (r networkConnectionMeshResource) GetConnectionsListByAgentID(ctx context.Context, agentIDs []int32) ([]Connection, error) {
	var filter []syntropy.V1AgentPairFilter
	for i := 0; i < len(agentIDs)-1; i++ {
		for j := i + 1; j < len(agentIDs); j++ {
			filter = append(filter, syntropy.V1AgentPairFilter{
				Agent2Id: agentIDs[i],
				Agent1Id: agentIDs[j],
			})
		}
	}

	connectionList, _, err := r.provider.client.ConnectionsApi.V1NetworkConnectionsSearch(ctx).V1NetworkConnectionsSearchRequest(syntropy.V1NetworkConnectionsSearchRequest{
		Filter: &syntropy.V1ConnectionFilter{
			AgentPair: filter,
		},
		Order: nil,
		Skip:  nil,
		Take:  nil,
	}).Execute()
	if err != nil {
		return nil, err
	}

	var connections []Connection
	for _, connection := range connectionList.Data {
		connections = append(connections, Connection{
			Agent1ID:          connection.Agent1.AgentId,
			Agent2ID:          connection.Agent2.AgentId,
			ConnectionGroupID: connection.AgentConnectionGroupId,
		})
	}
	return connections, nil
}

func (r networkConnectionMeshResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
