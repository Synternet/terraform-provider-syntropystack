data "syntropystack_agent" "agent1" {
  name = "syntropy-agent-prod-1"
}

data "syntropystack_agent" "agent2" {
  name = "syntropy-agent-prod-2"
}

resource "syntropystack_network_connection" "p2p" {
  agent_peer  = [data.syntropystack_agent.agent1.id, data.syntropystack_agent.agent2.id]
  sdn_enabled = false
}

data "syntropystack_network_connection_service" "svc" {
  connection_group_id = syntropystack_network_connection.p2p.id
  agent_id            = data.syntropystack_agent.agent1.id
  filter = {
    service_type = "DOCKER"
  }
}