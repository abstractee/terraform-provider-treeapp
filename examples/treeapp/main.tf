terraform {
  required_providers {
    treeapp = {
      source = "hashicorp.com/treeapp"
    }
  }
}

resource "treeapp" "myforest" {
  quantity  = 10
  frequency = "one_time"
}
