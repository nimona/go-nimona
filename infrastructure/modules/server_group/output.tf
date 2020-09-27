output "servers" {
  value = {
    for name, server in scaleway_instance_server.server :
    name => {
      id         = server.id
      name       = name
      user       = var.user
      ip_address = server.public_ip
      hostname   = cloudflare_record.server[name].hostname
    }
  }
}
