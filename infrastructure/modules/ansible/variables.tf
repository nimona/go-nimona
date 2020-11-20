variable "environment" {
  type = string
}

variable "servers" {
  type = list(map(string))
}

variable "volumes" {
  type = list(map(string))
}

variable "prometheus_jobs" {
  type = list(map(string))
}

variable "vault_password" {
  type = string
}

variable "ssh_private_key_file" {
  type    = string
  default = "ssh/id_rsa"
}

variable "nimona_version" {
  type    = string
  default = "latest"
}

variable "tags" {
  type    = string
  default = ""
}

variable "limit" {
  type    = string
  default = ""
}

variable "extra_args" {
  type    = string
  default = ""
}

variable "force_color" {
  type    = string
  default = "1"
}

variable "skip" {
  type    = bool
  default = false
}
