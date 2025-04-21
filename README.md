# Terraform Provider for TreeApp 🌳

**Terraform Provider TreeApp** is a custom [Terraform](https://terraform.io) provider written in Go that allows you to *literally terraform the Earth* by planting real trees via [TreeApp.com](https://treeapp.com).

Use infrastructure as code to contribute to reforestation efforts and make environmental impact as part of your CI/CD workflows or infrastructure deployments.

---

## 🌱 Features

- Plant trees as a one-time effort or recurring schedule
- Easily integrate tree planting into your Terraform workflows
- Automate your environmental impact with code

---


## 📦 Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

To use the provider, add it to your Terraform configuration:

```hcl
terraform {
  required_providers {
    treeapp = {
      source  = "hashicorp.com/test/treeapp" # use your own provider path
    }
  }
}
```

# 🔐 Authentication

You will need a TreeApp API key to authenticate. Pass it in via the provider block : 
```hcl
provider "treeapp" {
  api_key = var.treeapp_api_key
}

variable "treeapp_api_key" {}

```
# 🌳 Usage Example

Here's a minimal example that plants 6 trees as a one-time action:
```hcl
resource "treeapp_tree" "myforest" {
  quantity  = 6
  frequency = "one_time"
}
```

# Parameters
| Name     | Type   | Description                                        |
|----------|--------|----------------------------------------------------|
| quantity | int    | Number of trees to plant                           |
| frequency| string | How often to plant: one_time, per_deployment, per_month |
