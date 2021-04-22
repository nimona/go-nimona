variable "environment" {
  type    = string
  default = ""
}

variable "group" {
  type = string
}

variable "instances" {
  type = set(string)
}

variable "volumes" {
  type = map(any)
}

variable "ssh_private_key_file" {
  type = string
}

variable "ssh_public_key_file" {
  type = string
}

variable "cloudflare_zone_id" {
  type = string
}

variable "user" {
  type    = string
  default = "deploy"
}

variable "image" {
  type    = string
  default = ""
}

variable "type" {
  type    = string
  default = ""
}

variable "tags" {
  type    = list(string)
  default = []
}

variable "inbound_ports" {
  type    = set(number)
  default = []
}

variable "reverse_dns" {
  type    = bool
  default = false
}
