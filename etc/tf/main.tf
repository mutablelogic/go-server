terraform {
  required_providers {
    kaiak = {
      source = "mutablelogic/kaiak"
    }
  }
}

provider "kaiak" {
  # Defaults to KAIAK_ENDPOINT or http://localhost:8084/api
}

///////////////////////////////////////////////////////////////////////////////
// Data source — list available resource types

data "kaiak_resources" "all" {}

output "resource_types" {
  description = "All resource types known to the running server"
  value       = data.kaiak_resources.all.resources[*].name
}

///////////////////////////////////////////////////////////////////////////////
// Static file server

resource "kaiak_httpstatic" "docs" {
  name = "docs"
  path = "/docs"
  dir  = "/tmp/kaiak-docs"
}

///////////////////////////////////////////////////////////////////////////////
// Logger with debug enabled

resource "kaiak_logger" "debug" {
  name  = "debug"
  debug = true
}
