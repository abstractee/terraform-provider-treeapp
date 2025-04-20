terraform {
  required_providers {
    treeapp = {
      source = "hashicorp.com/test/treeapp"
    }
  }
}

resource "treeapp_tree" "myforest" {
  quantity  = 10
  frequency = "one_time"
}
