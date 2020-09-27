provider "local" {}
provider "null" {}

resource "local_file" "vault_password" {
  filename        = "${path.module}/.vault-password"
  file_permission = "0644"
  content         = var.vault_password
}

resource "local_file" "inventory" {
  filename = "${path.module}/inventories/${var.environment}"
  content = templatefile("${path.module}/templates/inventory.tpl", {
    server_groups = var.server_groups
  })
}

resource "null_resource" "run" {
  triggers = { always_run = "${timestamp()}" }

  depends_on = [
    local_file.vault_password,
    local_file.inventory
  ]

  provisioner "local-exec" {
    working_dir = path.module
    command     = "%{if var.skip}true%{else}make prepare%{endif}"
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
