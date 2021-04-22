output "servers" {
  value = flatten([
    for name, server in scaleway_instance_server.server :
    {
      id         = server.id
      name       = name
      user       = var.user
      ip_address = server.public_ip
      hostname   = cloudflare_record.server[name].hostname
    }
  ])
}

output "volumes" {
  value = flatten([
    for name, server in scaleway_instance_server.server :
    [
      for vol_name, attrs in scaleway_instance_volume.block :
      tomap({
        id         = attrs.id
        name       = local.volumes[vol_name].vol_name
        server     = name
        hostname   = cloudflare_record.server[name].hostname
        size_in_gb = attrs.size_in_gb
        mountpoint = local.volumes[vol_name].mountpoint
      }) if local.volumes[vol_name].server == name
    ]
  ])
}
