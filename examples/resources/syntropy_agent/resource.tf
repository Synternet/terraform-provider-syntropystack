resource "syntropy_agent" "agent" {
  name        = "terraform-provider-syntropystack-agent"
  provider_id = 2
  token       = "random-token"
}