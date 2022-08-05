resource "syntropystack_agent" "agent" {
  name        = "terraform-provider-syntropystack-agent"
  provider_id = 3
  token       = "<AGENT_TOKEN>"
}