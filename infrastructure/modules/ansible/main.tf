provider "local" {}
provider "null" {}

locals {
  servers_by_hostname = {
    for server in var.servers :
    server.hostname => server
  }
  server_groups = distinct([for server in var.servers : server.group])
  servers_by_group = {
    for group in local.server_groups :
    group => {
      for server in var.servers :
      server.name => server if server.group == group
    }
  }
  volumes_by_hostname = {
    for hostname in keys(local.servers_by_hostname) :
    hostname => {
      for volume in var.volumes :
      volume.name => volume if volume.hostname == hostname
    }
  }
}

resource "local_file" "vault_password" {
  filename        = "${path.module}/.vault-password"
  file_permission = "0644"
  content         = var.vault_password
}

resource "local_file" "inventory" {
  filename        = "${path.module}/inventories/${var.environment}"
  file_permission = "0644"
  content = templatefile("${path.module}/templates/inventory.tpl", {
    server_groups = tomap(local.servers_by_group)
  })
}

resource "local_file" "volumes" {
  for_each = local.volumes_by_hostname

  filename        = "${path.module}/host_vars/${each.key}/volumes.yml"
  file_permission = "0644"
  content = templatefile("${path.module}/templates/volumes.tpl", {
    volumes = {
      for name, vol in each.value :
      name => {
        id         = split("/", vol.id)[length(split("/", vol.id)) - 1]
        size_in_gb = tonumber(vol.size_in_gb)
        mountpoint = vol.mountpoint
      }
    }
  })
}

resource "local_file" "prometheus_jobs" {
  for_each = {
    for server in var.servers :
    server.name => server if server.group == "metrics"
  }

  filename = join("/", [
    path.module, "host_vars", each.value.hostname, "prometheus_jobs.yml"
  ])
  file_permission = "0644"
  content = templatefile("${path.module}/templates/prometheus_jobs.tpl", {
    servers_by_group = local.servers_by_group
    prometheus_jobs  = var.prometheus_jobs
  })
}

resource "null_resource" "run" {
  triggers = { always_run = timestamp() }

  depends_on = [
    local_file.vault_password,
    local_file.inventory
  ]

  provisioner "local-exec" {
    working_dir = path.module
    command     = <<CMD
%{if var.skip || var.skip_prepare}
true
%{else}
make prepare
%{endif}
CMD
  }

  provisioner "local-exec" {
    working_dir = path.module
    environment = { ANSIBLE_FORCE_COLOR = var.force_color }

    command = <<CMD
%{if var.skip~}
true
%{else~}
ansible-playbook \
  --inventory "${abspath(local_file.inventory.filename)}" \
  --private-key "${var.ssh_private_key_file}" \
  --extra-vars nimona_version="${trimspace(var.nimona_version)}" \
%{if var.limit != ""~}
  --limit "${var.limit}" \
%{endif~}
%{if var.tags != ""~}
  --tags "${var.tags}" \
%{endif~}
%{if var.extra_args != ""~}
  ${var.extra_args} \
%{endif~}
  site.yml
%{endif~}
CMD
  }
}
