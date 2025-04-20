terraform {
  required_providers {
    treeapp = {
      source = "hashicorp.com/test/treeapp"
    }
  }
}

resource "treeapp" "myforest" {
  quantity  = 10
  frequency = "one_time"
}
