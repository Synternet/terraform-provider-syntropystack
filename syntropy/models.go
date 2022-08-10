package syntropy

import "github.com/hashicorp/terraform-plugin-framework/types"

type Connection struct {
	Agent1ID          int32 `tfsdk:"agent_1_id"`
	Agent2ID          int32 `tfsdk:"agent_2_id"`
	ConnectionGroupID int32 `tfsdk:"agent_connection_group_id"`
}
type NetworkConnectionMeshData struct {
	ID          types.String `tfsdk:"id"`
	AgentIds    []int32      `tfsdk:"agent_ids"`
	Connections []Connection `tfsdk:"connections"`
	SdnEnabled  types.Bool   `tfsdk:"sdn_enabled"`
}

type NetworkConnectionData struct {
	ID         types.Int64 `tfsdk:"id"`
	AgentIds   []int64     `tfsdk:"agent_peer"`
	SdnEnabled types.Bool  `tfsdk:"sdn_enabled"`
}

type AgentResource struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ProviderId types.Int64  `tfsdk:"provider_id"`
	Token      types.String `tfsdk:"token"`
	Tags       []string     `tfsdk:"tags"`
}

type AgentSearchDataSource struct {
	Skip   types.Int64  `tfsdk:"skip"`
	Take   types.Int64  `tfsdk:"take"`
	Search types.String `tfsdk:"search"`
	Filter *AgentFilter `tfsdk:"filter"`
	Agents []Agent      `tfsdk:"agents"`
}

type Agent struct {
	ID              types.Int64    `tfsdk:"id"`
	Name            string         `tfsdk:"name"`
	PublicIPv4      types.String   `tfsdk:"public_ipv4"`
	Status          types.String   `tfsdk:"status"`
	IsOnline        types.Bool     `tfsdk:"is_online"`
	Version         types.String   `tfsdk:"version"`
	LocationCountry types.String   `tfsdk:"location_country"`
	LocationCity    types.String   `tfsdk:"location_city"`
	DeviceID        types.String   `tfsdk:"device_id"`
	IsVirtual       types.Bool     `tfsdk:"is_virtual"`
	Type            types.String   `tfsdk:"type"`
	ModifiedAt      types.String   `tfsdk:"modified_at"`
	Tags            []Tag          `tfsdk:"tags"`
	AgentProvider   *AgentProvider `tfsdk:"provider"`
}

type AgentFilter struct {
	ID              *[]int64  `tfsdk:"id"`
	Name            *string   `tfsdk:"name"`
	TagID           *[]int64  `tfsdk:"tag_id"`
	ProviderID      *[]int64  `tfsdk:"provider_id"`
	Type            *[]string `tfsdk:"type"`
	Version         *[]string `tfsdk:"version"`
	TagName         *[]string `tfsdk:"tag_name"`
	Status          *[]string `tfsdk:"status"`
	LocationCountry *[]string `tfsdk:"location_country"`
	ModifiedAtFrom  *string   `tfsdk:"modified_at_from"`
	ModifiedAtTo    *string   `tfsdk:"modified_at_to"`
}

type Tag struct {
	ID   int64  `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

type AgentProvider struct {
	ID   int64  `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

type NetworkConnectionServiceDataSource struct {
	ID                types.String                    `tfsdk:"id"`
	ConnectionGroupID int64                           `tfsdk:"connection_group_id"`
	AgentID           int64                           `tfsdk:"agent_id"`
	Filter            *NetworkConnectionServiceFilter `tfsdk:"filter"`
	Subnets           []ServiceSubnet                 `tfsdk:"subnets"`
}

type NetworkConnectionServiceFilter struct {
	ServiceName *string `tfsdk:"service_name"`
	ServiceType *string `tfsdk:"service_type"`
	SubnetID    *int64  `tfsdk:"subnet_id"`
}

type ServiceSubnet struct {
	ID      int64  `tfsdk:"subnet_id"`
	IP      string `tfsdk:"subnet_ip"`
	Enabled bool   `tfsdk:"is_subnet_enabled"`
}
