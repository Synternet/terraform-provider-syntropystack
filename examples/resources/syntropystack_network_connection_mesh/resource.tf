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