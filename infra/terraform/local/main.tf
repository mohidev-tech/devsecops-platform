terraform {
  required_version = ">= 1.6.0"
  required_providers {
    kind = {
      source  = "tehcyx/kind"
      version = "~> 0.4"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.13"
    }
  }
}

provider "kind" {}

resource "kind_cluster" "this" {
  name           = "devsecops"
  wait_for_ready = true

  # kind_config is a nested block (not an argument) in tehcyx/kind >= 0.4.
  # Three nodes total: one control-plane + two workers.
  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"
    }

    node {
      role = "worker"
    }

    node {
      role = "worker"
    }
  }
}

provider "helm" {
  kubernetes {
    config_path = kind_cluster.this.kubeconfig_path
  }
}

resource "helm_release" "ingress_nginx" {
  name             = "ingress-nginx"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  namespace        = "ingress-nginx"
  create_namespace = true
  version          = "4.10.1"
  depends_on       = [kind_cluster.this]
}

output "kubeconfig" {
  value = kind_cluster.this.kubeconfig_path
}
