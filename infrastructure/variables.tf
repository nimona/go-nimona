variable "node_names" {
  type    = set(string)
  default = ["asimov", "egan", "sloan"]
}

variable "node_server_user" {
  type    = string
  default = "deploy"
}

variable "node_server_image" {
  type    = string
  default = "ubuntu_focal"
}

variable "node_server_type" {
  type    = string
  default = "DEV1-S"
}

variable "node_server_tags" {
  type    = list(string)
  default = ["nimona", "node"]
}

variable "node_dns_suffix" {
  type    = string
  default = ".node"
}

variable "ssh_private_key_file" {
  type    = string
  default = "ssh/id_rsa"
}

variable "ssh_public_key_file" {
  type    = string
  default = "ssh/id_rsa.pub"
}

variable "cloudflare_zone_id" {
  type = string
}
