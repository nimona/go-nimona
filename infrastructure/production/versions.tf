terraform {
  required_providers {
    cloudflare = {
      source = "cloudflare/cloudflare"
    }
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
  required_version = ">= 0.13"
}
