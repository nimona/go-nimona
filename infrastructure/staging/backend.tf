terraform {
  backend "s3" {
    bucket                      = "nimona-staging-terraform"
    key                         = "nimona-staging.tfstate"
    region                      = "fr-par"
    endpoint                    = "https://s3.fr-par.scw.cloud"
    skip_credentials_validation = true
    skip_region_validation      = true
  }
}
