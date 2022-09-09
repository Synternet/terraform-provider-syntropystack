package syntropy

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = networkConnectionServiceDataSourceType{}
var _ tfsdk.DataSource = networkConnectionServiceDataSource{}

type networkConnectionServiceDataSourceType struct{}

func (d networkConnectionServiceDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Datasource retrieves list of services that were discovered in connection",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Network connection service ID randomly generated",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"connection_group_id": {
				Description: "Unique identifier for the connection.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"filter": {
				Description: "Network connection service filters",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"service_name_substring": {
						Description: "Filter service list by connection service name substring that is running on agent",
						Type:        types.StringType,
						Optional:    true,
					},
					"service_type": {
						Description: "Filter service list by connection service type that is running on agent",
						Type:        types.StringType,
						Optional:    true,
					},
					"service_id": {
						Description: "Filter service list by subnet ID",
						Type:        types.Int64Type,
						Optional:    true,
					},
					"agent_id": {
						Description: "Filter service list by agent ID",
						Type:        types.Int64Type,
						Optional:    true,
					},
				}),
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

func (d networkConnectionServiceDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return networkConnectionServiceDataSource{
		provider: provider,
	}, diags
}

type networkConnectionServiceDataSource struct {
	provider provider
}

func (d networkConnectionServiceDataSource) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	var data NetworkConnectionServiceDataSource
	ctx = d.provider.createAuthContext(ctx)
	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := d.provider.client.ConnectionsApi.V1NetworkConnectionsServicesGet(ctx).Filter(fmt.Sprint(data.ConnectionGroupID)).Execute()
	if err != nil {
		response.Diagnostics.AddError("Error while getting network connection services", err.Error())
		return
	}

	if len(resp.Data) != 1 {
		response.Diagnostics.AddError(fmt.Sprintf("Something went wrong. Expected 1 connection, but got %d", len(resp.Data)), fmt.Sprintf("connection = %s", fmt.Sprint(data.ConnectionGroupID)))
		return
	}

	connectionDetails, err := getOneConnectionDetails(ctx, *d.provider.client.ConnectionsApi, data.ConnectionGroupID)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Unable to get connection %v services", fmt.Sprint(data.ConnectionGroupID)), err.Error())
		return
	}

	var filteredServices []ConnectionServiceData
	// Filter results
	for _, svc := range connectionDetails.Services {
		// Filter by agent ID field
		if data.Filter != nil && !data.Filter.AgentID.IsNull() && data.Filter.AgentID.Value != svc.AgentID {
			continue
		}

		// Filter by service type field
		if data.Filter != nil && !data.Filter.ServiceType.IsNull() && data.Filter.ServiceType.Value != svc.Type {
			continue
		}

		// Filter by service name field
		if data.Filter != nil && !data.Filter.ServiceName.IsNull() && !strings.Contains(svc.Name, data.Filter.ServiceName.Value) {
			continue
		}

		// Filter by service ID field
		if data.Filter != nil && !data.Filter.ServiceID.IsNull() && data.Filter.ServiceID.Value != svc.ID {
			continue
		}

		filteredServices = append(filteredServices, svc)
	}

	data.ID = types.String{Value: uuid.New().String()}
	data.Services = filteredServices
	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}
