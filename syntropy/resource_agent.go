package syntropy

import (
	"context"
	"fmt"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = agentResourceType{}
var _ tfsdk.Resource = agentResource{}
var _ tfsdk.ResourceWithImportState = agentResource{}

type agentResourceType struct{}

type agentResource struct {
	provider provider
}

func (t agentResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Creates virtual Syntropy platform agent",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Agent ID",
				Type:        types.Int64Type,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"name": {
				Description: "Agent name",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"provider_id": {
				Description: "Agent provider ID",
				Type:        types.Int64Type,
				Required:    true,
			},
			"token": {
				Description: "Agent token",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"tags": {
				Description: "Agent tags",
				Optional:    true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
			},
		},
	}, nil
}

func (t agentResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return agentResource{
		provider: provider,
	}, diags
}

func (r agentResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan AgentData
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pId := int32(plan.ProviderId.Value)
	agent, _, err := r.provider.client.AgentsApi.V1NetworkAgentsCreate(ctx).V1NetworkAgentsCreateRequest(syntropy.V1NetworkAgentsCreateRequest{
		AgentName:       plan.Name.Value,
		AgentProviderId: &pId,
		AgentToken:      plan.Token.Value,
		AgentTags:       plan.Tags,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while creating virtual agent", err.Error())
		return
	}

	plan.ID = types.Int64{Value: int64(agent.Data.AgentId)}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r agentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state AgentData
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	agent, _, err := r.provider.client.AgentsApi.V1NetworkAgentsGet(ctx).Filter(state.ID.String()).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while getting virtual agent", err.Error())
		return
	}

	if len(agent.Data) != 1 {
		resp.Diagnostics.AddError("Something went wrong getting virtual agent", fmt.Sprintf("Agent count %d, but expected 1", len(agent.Data)))
		return
	}

	var tags []string
	for _, tag := range agent.Data[0].AgentTags {
		tags = append(tags, tag.AgentTagName)
	}

	state.Name = types.String{Value: agent.Data[0].AgentName}
	state.ProviderId = types.Int64{Value: int64(agent.Data[0].AgentProvider.AgentProviderId)}
	state.Tags = tags

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r agentResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan AgentData
	ctx = r.provider.createAuthContext(ctx)
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pId := int32(plan.ProviderId.Value)
	_, err := r.provider.client.AgentsApi.V1NetworkAgentsUpdate(ctx, int32(plan.ID.Value)).V1NetworkAgentsUpdateRequest(syntropy.V1NetworkAgentsUpdateRequest{
		AgentTags:       plan.Tags,
		AgentName:       &plan.Name.Value,
		AgentProviderId: &pId,
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while updating virtual agent", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r agentResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data AgentData
	ctx = r.provider.createAuthContext(ctx)
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.provider.client.AgentsApi.V1NetworkAgentsRemove(ctx).V1NetworkAgentsRemoveRequest(syntropy.V1NetworkAgentsRemoveRequest{
		AgentIds: []int32{int32(data.ID.Value)},
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error while deleting virtual agent", err.Error())
		return
	}
}

func (r agentResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
