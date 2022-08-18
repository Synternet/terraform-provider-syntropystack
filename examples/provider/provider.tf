terraform {
  required_providers {
    syntropystack = {
      source  = "SyntropyNet/syntropystack"
      version = "~> 0.1"
    }
  }
}

provider "syntropystack" {
  access_token = "<ACCESS_TOKEN>"
}