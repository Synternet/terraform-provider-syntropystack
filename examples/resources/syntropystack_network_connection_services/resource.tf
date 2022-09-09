resource "syntropystack_network_connection" "p2p" {
  agent_peer  = [1, 2]
  sdn_enabled = false
}

data "syntropystack_network_connection_services" "filtered_services" {
  connection_group_id = syntropystack_network_connection.p2p.id
  filter = {
    service_name_substring = "movie-service"
  }
}

resource "syntropystack_network_connection_services" "enabled_services" {
  connection_group_id = syntropystack_network_connection.p2p.id

  services = [
    for svc in data.syntropystack_network_connection_services.filtered_services.services : {
      id      = svc.id
      enabled = true
    }
  ]
}
