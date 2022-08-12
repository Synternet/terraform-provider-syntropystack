package syntropy

import (
	"context"
	"fmt"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
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
				Description: "Network connection group ID",
				Type:        types.Int64Type,
				Required:    true,
			},
			"agent_id": {
				Description: "Syntropy agent ID",
				Type:        types.Int64Type,
				Required:    true,
			},
			"filter": {
				Description: "Network connection service filters",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"service_name": {
						Description: "Filter service list by connection service name that is running on agent",
						Type:        types.StringType,
						Optional:    true,
					},
					"service_type": {
						Description: "Filter service list by connection service type that is running on agent",
						Type:        types.StringType,
						Optional:    true,
					},
					"subnet_id": {
						Description: "Filter service list by subnet ID",
						Type:        types.Int64Type,
						Optional:    true,
					},
				}),
			},
			"subnets": {
				Description: "List of subnets discovered in services",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"subnet_id": {
						Description: "Network connection service subnet ID",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"subnet_ip": {
						Description: "Network connection service subnet IP",
						Type:        types.StringType,
						Computed:    true,
					},
					"is_subnet_enabled": {
						Description: "Is network connection service subnet enabled",
						Type:        types.BoolType,
						Computed:    true,
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

	csID := strconv.FormatInt(data.ConnectionGroupID, 10)
	resp, _, err := d.provider.client.ConnectionsApi.V1NetworkConnectionsServicesGet(ctx).Filter(csID).Execute()
	if err != nil {
		response.Diagnostics.AddError("Error while getting network connection services", err.Error())
		return
	}

	if len(resp.Data) == 0 {
		response.Diagnostics.AddError(fmt.Sprintf("Connection not found by ID = %s", csID), "")
		return
	}

	var services []syntropy.V1ConnectionServiceAgentService

	if resp.Data[0].Agent1.AgentId == int32(data.AgentID) {
		services = resp.Data[0].Agent1.AgentServices
	} else {
		services = resp.Data[0].Agent2.AgentServices
	}

	// Loop though connection services
	for _, service := range services {

		// Filter by agent service name field
		if data.Filter != nil && data.Filter.ServiceName != nil && *data.Filter.ServiceName != service.AgentServiceName {
			continue
		}

		// Filter by agent service type field
		if data.Filter != nil && data.Filter.ServiceType != nil && *data.Filter.ServiceType != string(service.AgentServiceType) {
			continue
		}

		// Loop through service subnets
		for _, subnet := range service.AgentServiceSubnets {

			// Filter by subnet ID field
			if data.Filter != nil && data.Filter.SubnetID != nil && *data.Filter.SubnetID != int64(subnet.AgentServiceSubnetId) {
				continue
			}

			subnetEnabled := false

			// Loop through remote subnets which were enabled
			for _, enabledSubnet := range resp.Data[0].AgentConnectionSubnets {
				if subnet.AgentServiceSubnetId == enabledSubnet.AgentServiceSubnetId {
					subnetEnabled = enabledSubnet.AgentConnectionSubnetIsEnabled
				}
			}
			data.Subnets = append(data.Subnets, ServiceSubnet{
				ID:      int64(subnet.AgentServiceSubnetId),
				IP:      subnet.AgentServiceSubnetIp,
				Enabled: subnetEnabled,
			})
		}
	}
	data.ID = types.String{Value: uuid.New().String()}
	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}
