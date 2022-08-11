---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "syntropystack_network_connection_mesh Resource - terraform-provider-syntropystack"
subcategory: ""
description: |-
  Creates network mesh between agents
---

# syntropystack_network_connection_mesh (Resource)

Creates network mesh between agents

## Example Usage

```terraform
data "syntropystack_agent_search" "results" {
  filter = {
    type             = ["LINUX"]
    location_country = ["US", "UK"]
  }
}

resource "syntropystack_network_connection_mesh" "test_connection_mesh" {
  agent_ids   = data.syntropystack_agent_search.results.agents.*.id
  sdn_enabled = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `agent_ids` (Set of Number) List of agent IDs for network connection mesh

### Optional

- `sdn_enabled` (Boolean) Should SDN be enabled?

### Read-Only

- `connections` (Attributes Set) Created connections (see [below for nested schema](#nestedatt--connections))
- `id` (String) Network connection mesh ID randomly generated

<a id="nestedatt--connections"></a>
### Nested Schema for `connections`

Read-Only:

- `agent_1_id` (Number) Agent 1 ID
- `agent_2_id` (Number) Agent 2 ID
- `agent_connection_group_id` (Number) Agent connection group ID

