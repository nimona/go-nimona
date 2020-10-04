locals {
  image    = coalesce(var.image, "ubuntu_focal")
  type     = coalesce(var.type, "DEV1-S")
  dns_root = var.environment == "production" ? "" : var.environment
}

resource "scaleway_instance_ip" "server" {
  for_each = var.instances
}

resource "cloudflare_record" "server" {
  for_each = scaleway_instance_ip.server
  zone_id  = var.cloudflare_zone_id

  name = join(".", compact([
    each.key == var.group ? "" : each.key,
    var.group,
    local.dns_root
  ]))

  value = each.value.address
  type  = "A"
}

resource "cloudflare_record" "wildcard" {
  for_each = scaleway_instance_ip.server
  zone_id  = var.cloudflare_zone_id

  name = join(".", compact([
    "*",
    each.key == var.group ? "" : each.key,
    var.group,
    local.dns_root
  ]))

  value = each.value.address
  type  = "A"
}

resource "scaleway_instance_security_group" "server" {
  name                    = var.group
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"

  inbound_rule {
    action = "accept"
    port   = "22"
  }

  inbound_rule {
    action = "accept"
    port   = "80"
  }

  inbound_rule {
    action = "accept"
    port   = "443"
  }

  dynamic "inbound_rule" {
    for_each = var.inbound_ports

    content {
      action = "accept"
      port   = inbound_rule.value
    }
  }
}

resource "scaleway_instance_server" "server" {
  for_each = scaleway_instance_ip.server

  name = join("-", compact([
    var.group,
    each.key != var.group ? each.key : ""
  ]))

  type              = local.type
  image             = local.image
  tags              = concat([var.environment, var.group], var.tags)
  ip_id             = each.value.id
  security_group_id = scaleway_instance_security_group.server.id

  # initialization sequence
  cloud_init = templatefile("${path.module}/cloud-init.tpl", {
    user           = var.user,
    ssh_public_key = file(var.ssh_public_key_file)
  })

  connection {
    host        = self.public_ip
    user        = var.user
    private_key = file(var.ssh_private_key_file)
  }

  provisioner "remote-exec" {
    inline = [
      "tail -f /var/log/cloud-init-output.log &",
      "while [ ! -f /var/lib/cloud/instance/boot-finished ]; do sleep 2; done;"
    ]
  }

  # Wait enough time that next remote-exec doesn't manage to connect before
  # shutdown, but instead retries until the machine comes back up.
  provisioner "local-exec" {
    command = "sleep 30"
  }

  # Ensure machines are back up
  provisioner "remote-exec" {
    inline = [
      "while [ ! -f /var/lib/cloud/instance/boot-finished ]; do sleep 2; done;"
    ]
  }
}

resource "scaleway_instance_ip_reverse_dns" "base" {
  for_each = var.reverse_dns ? scaleway_instance_server.server : {}

  ip_id   = scaleway_instance_ip.server[each.key].id
  reverse = cloudflare_record.server[each.key].hostname
}
