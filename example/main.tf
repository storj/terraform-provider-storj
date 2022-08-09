terraform {
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = "~> 2.0"
    }

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

resource "storj_bucket" "app_bucket" {
  bucket = "app"
}

resource "storj_access_grant" "app_grant" {
  access_grant = var.root_access_grant

  bucket {
    name = storj_bucket.app_bucket.bucket
    paths = []
  }
}

resource "helm_release" "app" {
  repository = "https://mjpitz.com"
  chart = "12factor"
  name  = "app"

  set {
    name  = "deployment.application.env[0].name"
    value = "STORJ_ACCESS_GRANT"
  }

  set_sensitive {
    name  = "deployment.application.env[0].value"
    value = storj_access_grant.app_grant.derived_access_grant
  }
}
