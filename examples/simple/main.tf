terraform {
  required_providers {
    storj = {
      source  = "terraform.local/local/storj"
      version = "1.0.0"
    }
  }
}

provider "storj" {}

variable "root_access_grant" {
  type = string
}

resource "storj_bucket" "terraform_bucket" {
  bucket = "terraform-testing"
}

resource "storj_access_grant" "terraform_grant" {
  access_grant = var.root_access_grant

  bucket {
    name = storj_bucket.terraform_bucket.bucket
    paths = []
  }

  allow_download = true
  allow_upload = true
  allow_list = true
  allow_delete = true
}

resource "storj_object" "terraform_state" {
  bucket = storj_bucket.terraform_bucket.bucket
  key = "main.tf"

  source = "main.tf"
  metadata = {
    "key1" = "value1"
    "key2" = "value2"
  }
}

output "terraform_bucket_grant" {
  value = storj_access_grant.terraform_grant.derived_access_grant
  sensitive = true
}
