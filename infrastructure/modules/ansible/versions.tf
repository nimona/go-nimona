terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 1.4"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 2.1"
    }
  }
  required_version = ">= 0.13"
}
