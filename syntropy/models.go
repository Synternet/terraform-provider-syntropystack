package syntropy

import "github.com/hashicorp/terraform-plugin-framework/types"

type Connection struct {
	Agent1ID          int32                   `tfsdk:"agent_1_id"`
	Agent2ID          int32                   `tfsdk:"agent_2_id"`
	ConnectionGroupID int32                   `tfsdk:"connection_group_id"`
	Services          []ConnectionServiceData `tfsdk:"services"`
}

type NetworkConnectionMeshEdit struct {
	ID          types.String `tfsdk:"id"`
	AgentIds    []int32      `tfsdk:"agent_ids"`
	Connections types.Set    `tfsdk:"connections"`
	SdnEnabled  types.Bool   `tfsdk:"sdn_enabled"`
}

type NetworkConnectionMesh struct {
	ID          types.String `tfsdk:"id"`
	AgentIds    []int32      `tfsdk:"agent_ids"`
	Connections []Connection `tfsdk:"connections"`
	SdnEnabled  types.Bool   `tfsdk:"sdn_enabled"`
}

type NetworkConnection struct {
	ID         types.Int64             `tfsdk:"id"`
	AgentIds   []int64                 `tfsdk:"agent_peer"`
	SdnEnabled types.Bool              `tfsdk:"sdn_enabled"`
	Services   []ConnectionServiceData `tfsdk:"services"`
}

type AgentResource struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
	Tags  []string     `tfsdk:"tags"`
}

type AgentSearchDataSource struct {
	Skip   types.Int64  `tfsdk:"skip"`
	Take   types.Int64  `tfsdk:"take"`
	Search types.String `tfsdk:"search"`
	Filter *AgentFilter `tfsdk:"filter"`
	Agents []AgentData  `tfsdk:"agents"`
}

type AgentData struct {
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

type Tag struct {
	ID   int64  `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

type AgentProvider struct {
	ID   int64  `tfsdk:"id"`
	Name string `tfsdk:"name"`
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

type NetworkConnectionServiceDataSource struct {
	ID                types.String                    `tfsdk:"id"`
	ConnectionGroupID int32                           `tfsdk:"connection_group_id"`
	Filter            *NetworkConnectionServiceFilter `tfsdk:"filter"`
	Services          []ConnectionServiceData         `tfsdk:"services"`
}

type NetworkConnectionServiceFilter struct {
	ServiceName types.String `tfsdk:"service_name_substring"`
	ServiceType types.String `tfsdk:"service_type"`
	ServiceID   types.Int64  `tfsdk:"service_id"`
	AgentID     types.Int64  `tfsdk:"agent_id"`
}

type ConnectionServiceData struct {
	ID           int64  `tfsdk:"id"`
	Name         string `tfsdk:"name"`
	IP           string `tfsdk:"ip"`
	Type         string `tfsdk:"type"`
	Enabled      bool   `tfsdk:"enabled"`
	AgentID      int64  `tfsdk:"agent_id"`
	ConnectionId int64  `tfsdk:"-"`
}

type ConnectionService struct {
	ConnectionGroupID types.Int64 `tfsdk:"connection_group_id"`
	Services          []Service   `tfsdk:"services"`
}

type Service struct {
	ID      int64 `tfsdk:"id"`
	Enabled bool  `tfsdk:"enabled"`
}
