resource "syntropy_network_connection_mesh" "test_connection_mesh" {
  agent_ids   = [1, 2, 3]
  sdn_enabled = false
}