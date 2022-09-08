package syntropy

import (
	"context"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = agentSearchDataSourceType{}
var _ tfsdk.DataSource = agentSearchDataSource{}

type agentSearchDataSourceType struct{}

func (d agentSearchDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Datasource retrieves Syntropy agent data list",
		Attributes: map[string]tfsdk.Attribute{
			"skip": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "Number of items to skip",
			},
			"take": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "Number of items to take",
			},
			"search": {
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Agent name pattern. This will be used to filter out agent names that doesn't have specified patter",
			},
			"filter": {
				Description: "Syntropy agent search filter",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Filter by agent ID",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.Int64Type,
						},
					},
					"tag_id": {
						Description: "Filter by agent tag ID",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.Int64Type,
						},
					},
					"provider_id": {
						Description: "Filter by agent provider ID",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.Int64Type,
						},
					},
					"type": {
						Description: "Filter by agent type",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"version": {
						Description: "Filter by agent version",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"tag_name": {
						Description: "Filter by agent tag name",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"status": {
						Description: "Filter by agent status",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"location_country": {
						Description: "Filter by agent location country",
						Optional:    true,
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
					"modified_at_from": {
						Description: "Filter by agent modified at from date",
						Optional:    true,
						Type:        types.StringType,
					},
					"modified_at_to": {
						Description: "Filter by agent modified at to date",
						Optional:    true,
						Type:        types.StringType,
					},
					"name": {
						Description: "Filter by agent name",
						Optional:    true,
						Type:        types.StringType,
					},
				}),
			},
			"agents": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "Unique identifier for the agent",
						Computed:    true,
						Type:        types.Int64Type,
					},
					"name": {
						Description: "Name of the agent as it appears in Platform UI",
						Computed:    true,
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
				}),
			},
		},
	}, nil
}

func (d agentSearchDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return agentSearchDataSource{
		provider: provider,
	}, diags
}

type agentSearchDataSource struct {
	provider provider
}

func (d agentSearchDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data AgentSearchDataSource
	ctx = d.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	agentFilter := &syntropy.V1AgentFilter{}

	if data.Filter != nil {
		filter, err := flattenAgentFilter(*data.Filter)
		if err != nil {
			resp.Diagnostics.AddError("Error while parsing agent filter data", err.Error())
			return
		}
		agentFilter = filter
	}

	skip := int32(data.Skip.Value)
	take := int32(data.Take.Value)
	aResp, _, err := d.provider.client.AgentsApi.V1NetworkAgentsSearch(ctx).V1NetworkAgentsSearchRequest(syntropy.V1NetworkAgentsSearchRequest{
		Filter: agentFilter,
		Order:  nil,
		Skip:   &skip,
		Take:   &take,
		Search: &data.Search.Value,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while getting Syntropy agent", err.Error())
		return
	}

	for _, agent := range aResp.Data {
		data.Agents = append(data.Agents, Agent{
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
		})
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func flattenAgentFilter(in AgentFilter) (*syntropy.V1AgentFilter, error) {
	out := &syntropy.V1AgentFilter{
		AgentName: in.Name,
	}

	if in.ID != nil {
		out.AgentId = int64ArrayToInt32Array(*in.ID)
	}

	if in.ProviderID != nil {
		out.AgentProviderId = int64ArrayToInt32Array(*in.ProviderID)
	}

	if in.TagID != nil {
		out.AgentTagId = int64ArrayToInt32Array(*in.TagID)
	}

	if in.Type != nil {
		out.AgentType = stringArrayToAgentTypeArray(*in.Type)
	}

	if in.Version != nil {
		out.AgentVersion = *in.Version
	}

	if in.TagName != nil {
		out.AgentTagName = *in.TagName
	}

	if in.LocationCountry != nil {
		out.AgentLocationCountry = *in.LocationCountry
	}

	if in.Status != nil {
		out.AgentStatus = stringArrayToAgentStatusArray(*in.Status)
	}

	if in.ModifiedAtFrom != nil {
		modifiedAtFrom, err := tfValueToDateP(*in.ModifiedAtFrom)
		if err != nil {
			return nil, err
		}
		out.AgentModifiedAtFrom = modifiedAtFrom
	}

	if in.ModifiedAtTo != nil {
		modifiedAtTo, err := tfValueToDateP(*in.ModifiedAtTo)
		if err != nil {
			return nil, err
		}
		out.AgentModifiedAtTo = modifiedAtTo
	}
	return out, nil
}
