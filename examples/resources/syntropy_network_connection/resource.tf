resource "syntropy_network_connection" "p2p" {
  agent_peer  = [1, 2]
  sdn_enabled = true
}