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
