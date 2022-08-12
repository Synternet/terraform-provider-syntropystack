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
		Description: "Datasource retrieves Syntropy agent data",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Network connection ID",
				Type:        types.Int64Type,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"name": {
				Description: "Filter by agent modified at to date",
				Required:    true,
				Type:        types.StringType,
			},
			"public_ipv4": {
				Description: "Agent public IP",
				Type:        types.StringType,
				Computed:    true,
			},
			"status": {
				Description: "Agent status",
				Type:        types.StringType,
				Computed:    true,
			},
			"is_online": {
				Description: "Agent online status",
				Type:        types.BoolType,
				Computed:    true,
			},
			"version": {
				Description: "Agent version",
				Type:        types.StringType,
				Computed:    true,
			},
			"location_country": {
				Description: "Agent location country code",
				Type:        types.StringType,
				Computed:    true,
			},
			"location_city": {
				Description: "Agent city location",
				Type:        types.StringType,
				Computed:    true,
			},
			"device_id": {
				Description: "Agent device id",
				Type:        types.StringType,
				Computed:    true,
			},
			"is_virtual": {
				Description: "Is agent virtual",
				Type:        types.BoolType,
				Computed:    true,
			},
			"type": {
				Description: "Agent type",
				Type:        types.StringType,
				Computed:    true,
			},
			"modified_at": {
				Description: "Agent modified date",
				Type:        types.StringType,
				Computed:    true,
			},
			"tags": {
				Description: "Agent tags",
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
				Description: "Agent provider details",
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
