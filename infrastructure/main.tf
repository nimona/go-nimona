locals {
  node_names = sort(var.node_names)
  node_count = length(var.node_names)
}

provider "scaleway" {
  zone   = "fr-par-1"
  region = "fr-par"
}

provider "cloudflare" {
  version = "~> 2.0"
}

resource "scaleway_instance_ip" "node" {
  count = local.node_count
}

resource "cloudflare_record" "node" {
  count   = local.node_count
  zone_id = var.cloudflare_zone_id
  name    = "${local.node_names[count.index]}${var.node_dns_suffix}"
  value   = scaleway_instance_ip.node[count.index].address
  type    = "A"
}

resource "cloudflare_record" "node-wildcard" {
  count   = local.node_count
  zone_id = var.cloudflare_zone_id
  name    = "*.${local.node_names[count.index]}${var.node_dns_suffix}"
  value   = scaleway_instance_ip.node[count.index].address
  type    = "A"
}

resource "scaleway_instance_security_group" "node" {
  name                    = "nimona-node"
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
}

resource "scaleway_instance_server" "node" {
  count             = local.node_count
  name              = "node-${local.node_names[count.index]}"
  type              = var.node_server_type
  image             = var.node_server_image
  tags              = var.node_server_tags
  ip_id             = scaleway_instance_ip.node[count.index].id
  security_group_id = scaleway_instance_security_group.node.id

  # initialization sequence
  cloud_init = templatefile("${path.module}/terraform/cloud-init.tpl", {
    user           = var.node_server_user,
    ssh_public_key = file(pathexpand(var.ssh_public_key_file))
  })

  connection {
    host        = scaleway_instance_ip.node[count.index].address
    user        = var.node_server_user
    private_key = file(pathexpand(var.ssh_private_key_file))
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
