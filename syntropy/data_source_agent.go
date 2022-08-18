package syntropy

import (
	"context"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = agentDataSourceType{}
var _ tfsdk.DataSource = agentDataSource{}

type agentDataSourceType struct{}

func (d agentDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Datasource retrieves Syntropy agent data by agent name",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier for the agent",
				Type:        types.Int64Type,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"name": {
				Description: "Name of the agent as it appears in Platform UI",
				Required:    true,
				Type:        types.StringType,
			},
			"public_ipv4": {
				Description: "IP address of the agent in IPv4 format",
				Type:        types.StringType,
				Computed:    true,
			},
			"status": {
				Description: "Current status of the agent.",
				Type:        types.StringType,
				Computed:    true,
			},
			"is_online": {
				Description: "Current status of the agent.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"version": {
				Description: "Version of the agent.",
				Type:        types.StringType,
				Computed:    true,
			},
			"location_country": {
				Description: "Agent's location country two-letter code.",
				Type:        types.StringType,
				Computed:    true,
			},
			"location_city": {
				Description: "City, where your agent is based",
				Type:        types.StringType,
				Computed:    true,
			},
			"device_id": {
				Description: "A unique agent identifier. Usually machine id or other unique UUID with a workspace id prefix (to scope this agent to workspace).",
				Type:        types.StringType,
				Computed:    true,
			},
			"is_virtual": {
				Description: "Indicates if it's a virtual agent.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"type": {
				Description: "Possible types: LINUX, MACOS, WINDOWS, VIRTUAL",
				Type:        types.StringType,
				Computed:    true,
			},
			"modified_at": {
				Description: "Date and time when this agent was modified. Formatted as an ISO 8601 date time string.",
				Type:        types.StringType,
				Computed:    true,
			},
			"tags": {
				Description: "Agent specific words that can help you to create some rules around specific tags.",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Agent tag id",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"name": {
						Description: "Agent tag name",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
			"provider": {
				Description: "Returns provider of agent's endpoint",
				Computed:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Agent provider id",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"name": {
						Description: "Agent provider name",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
		},
	}, nil
}

func (d agentDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return agentDataSource{
		provider: provider,
	}, diags
}

type agentDataSource struct {
	provider provider
}

func (d agentDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data Agent
	ctx = d.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	skip := int32(0)
	take := int32(1)
	aResp, _, err := d.provider.client.AgentsApi.V1NetworkAgentsSearch(ctx).V1NetworkAgentsSearchRequest(syntropy.V1NetworkAgentsSearchRequest{
		Filter: nil,
		Order:  nil,
		Skip:   &skip,
		Take:   &take,
		Search: &data.Name,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while getting Syntropy agent", err.Error())
		return
	}

	for _, agent := range aResp.Data {
		data = Agent{
			ID:              types.Int64{Value: int64(agent.AgentId)},
			Name:            agent.AgentName,
			PublicIPv4:      types.String{Value: agent.AgentPublicIpv4},
			Status:          types.String{Value: NullableAgentStatusToString(agent.AgentStatus)},
			IsOnline:        types.Bool{Value: agent.AgentIsOnline},
			Version:         types.String{Value: agent.AgentVersion},
			LocationCountry: types.String{Value: NullableStringToString(agent.AgentLocationCountry)},
			LocationCity:    types.String{Value: NullableStringToString(agent.AgentLocationCity)},
			DeviceID:        types.String{Value: agent.AgentDeviceId},
			IsVirtual:       types.Bool{Value: agent.AgentIsVirtual},
			Type:            types.String{Value: string(agent.AgentType)},
			ModifiedAt:      types.String{Value: agent.AgentModifiedAt.String()},
			Tags:            convertAgentTagsToTfValue(agent.AgentTags),
			AgentProvider: &AgentProvider{
				ID:   int64(agent.AgentProvider.AgentProviderId),
				Name: agent.AgentProvider.AgentProviderName,
			},
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
