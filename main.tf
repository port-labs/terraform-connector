terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    port-labs = {
      source = "port-labs/port-labs"
      version = "~> 0.4.6"
    }
  }
  backend "s3" {
    bucket = "tf-wrapper"
    key    = "{{ .storage_key }}"
    region = "eu-west-1"
  }
}

provider "aws" {
  region = "eu-west-1"
}

provider "port-labs" {}

variable "blueprint" {
  type = string
  description = "identifier of blueprint"
}

variable "entity_identifier" {
  type = string
  description = "identifier of entity"
}

variable "run_id" {
  type = string
  description = "identifier of action run"
}