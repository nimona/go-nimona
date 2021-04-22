variable "ansible_vault_password" {
  type = string
}

variable "cloudflare_zone_id" {
  type = string
}

variable "nimona_version" {
  type = string
}

variable "environment" {
  type    = string
  default = ""
}

variable "reverse_dns" {
  type    = bool
  default = false
}

variable "server_groups_file" {
  type    = string
  default = "server_groups.yml"
}

variable "ssh_private_key_file" {
  type    = string
  default = ""
}

variable "ssh_public_key_file" {
  type    = string
  default = ""
}

variable "ansible_tags" {
  type    = string
  default = ""
}

variable "ansible_limit" {
  type    = string
  default = ""
}

variable "ansible_extra_args" {
  type    = string
  default = ""
}

variable "ansible_force_color" {
  type    = string
  default = "1"
}

variable "ansible_skip" {
  type    = bool
  default = false
}

variable "ansible_skip_prepare" {
  type    = bool
  default = false
}
