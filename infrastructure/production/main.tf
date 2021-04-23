provider "cloudflare" {}

provider "scaleway" {
  zone   = "fr-par-1"
  region = "fr-par"
}

locals {
  environment   = coalesce(var.environment, basename(abspath(path.root)))
  server_groups = yamldecode(file(var.server_groups_file))

  ssh_private_key_file = abspath(
    coalesce(
      pathexpand(var.ssh_private_key_file),
      "${path.module}/ssh/id_rsa"
    )
  )
  ssh_public_key_file = abspath(
    coalesce(
      pathexpand(var.ssh_public_key_file),
      "${path.module}/ssh/id_rsa.pub"
    )
  )
}

module "server_groups" {
  source   = "../modules/server_group"
  for_each = local.server_groups

  environment   = local.environment
  group         = each.key
  instances     = each.value.instances
  volumes       = lookup(each.value, "volumes", {})
  type          = lookup(each.value, "type", "")
  image         = lookup(each.value, "image", "")
  tags          = lookup(each.value, "tags", [])
  inbound_ports = lookup(each.value, "inbound_ports", [])
  reverse_dns   = var.reverse_dns

  cloudflare_zone_id   = var.cloudflare_zone_id
  ssh_private_key_file = local.ssh_private_key_file
  ssh_public_key_file  = local.ssh_public_key_file
}

module "ansible" {
  source = "../modules/ansible"

  environment          = local.environment
  vault_password       = var.ansible_vault_password
  ssh_private_key_file = local.ssh_private_key_file
  nimona_version       = var.nimona_version
  limit                = var.ansible_limit
  tags                 = var.ansible_tags
  extra_args           = var.ansible_extra_args
  force_color          = var.ansible_force_color
  skip                 = var.ansible_skip
  skip_prepare         = var.ansible_skip_prepare

  servers = flatten([
    for name, group in module.server_groups :
    [
      for server in group.servers :
      merge({ group = name }, server)
    ]
  ])

  volumes = flatten([
    for name, group in module.server_groups :
    [
      for volume in group.volumes :
      merge({ group = name }, volume)
    ]
  ])

  prometheus_jobs = flatten([
    for group, details in local.server_groups :
    [
      for job in details.prometheus_jobs :
      merge({ group = group }, job)
    ] if length(lookup(details, "prometheus_jobs", [])) > 0
  ])
}
