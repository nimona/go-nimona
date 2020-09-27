terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 2.11"
    }
    scaleway = {
      source  = "scaleway/scaleway"
      version = "~> 1.16"
    }
  }
  required_version = ">= 0.13"
}
