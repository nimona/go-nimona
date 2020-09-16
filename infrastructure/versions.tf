terraform {
  required_providers {
    cloudflare = {
      source = "terraform-providers/cloudflare"
    }
    local = {
      source = "hashicorp/local"
    }
    scaleway = {
      source = "terraform-providers/scaleway"
    }
  }
  required_version = ">= 0.13"
}
