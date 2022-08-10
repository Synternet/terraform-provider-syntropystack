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

type AgentData struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ProviderId types.Int64  `tfsdk:"provider_id"`
	Token      types.String `tfsdk:"token"`
	Tags       []string     `tfsdk:"tags"`
}

type AgentDataSource struct {
	Skip   types.Int64  `tfsdk:"skip"`
	Take   types.Int64  `tfsdk:"take"`
	Search types.String `tfsdk:"search"`
	Filter *AgentFilter `tfsdk:"filter"`
	Agents []Agent      `tfsdk:"agents"`
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

type Agent struct {
	ID              int64         `tfsdk:"id"`
	Name            string        `tfsdk:"name"`
	PublicIPv4      string        `tfsdk:"public_ipv4"`
	Status          string        `tfsdk:"status"`
	IsOnline        bool          `tfsdk:"is_online"`
	Version         string        `tfsdk:"version"`
	LocationCountry string        `tfsdk:"location_country"`
	LocationCity    string        `tfsdk:"location_city"`
	DeviceID        string        `tfsdk:"device_id"`
	IsVirtual       bool          `tfsdk:"is_virtual"`
	Type            string        `tfsdk:"type"`
	ModifiedAt      string        `tfsdk:"modified_at"`
	Tags            []Tag         `tfsdk:"tags"`
	AgentProvider   AgentProvider `tfsdk:"provider"`
}

type Tag struct {
	ID   int64  `tfsdk:"id"`
	Name string `tfsdk:"name"`
}

type AgentProvider struct {
	ID   int64  `tfsdk:"id"`
	Name string `tfsdk:"name"`
}
