terraform {
  required_providers {
    treeapp = {
      source = "hashicorp.com/test/treeapp"
    }
  }
}

provider "treeapp" {
  api_key = var.treeapp_api_key
}

variable "treeapp_api_key" {}

resource "treeapp_tree" "myforest" {
  quantity  = 4
  frequency = "per_month"
}
