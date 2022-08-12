data "syntropystack_agent" "agent1" {
  name = "syntropy-agent-prod"
}

data "syntropystack_network_connection_service" "agent1_services" {
  connection_group_id = 1 # Connection ID
  agent_id            = data.syntropystack_agent.agent1.id
  filter = {
    service_type = "DOCKER"
  }
}

// Enable all docker services for syntropy-agent-prod agent on specified connection
resource "syntropystack_enable_network_connection_subnet" "agent1_services" {
  for_each = { for i, v in data.syntropystack_network_connection_service.agent1_services.subnets : i => v }

  connection_group_id = 1 # Connection ID
  subnet_id           = each.value.subnet_id
  enable              = true
}
