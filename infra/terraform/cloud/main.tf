provider "aws" {
  region = var.region
  default_tags {
    tags = {
      lifecycle   = "ephemeral"
      owner       = var.owner
      project     = "devsecops-platform"
      managed_by  = "terraform"
    }
  }
}

resource "random_id" "suffix" {
  byte_length = 2
}

locals {
  name = "${var.cluster_name}-${random_id.suffix.hex}"
  azs  = slice(data.aws_availability_zones.available.names, 0, 2)
}

data "aws_availability_zones" "available" {
  state = "available"
}

# ---------------------------------------------------------------------------
# VPC — public subnets only. Nodes get public IPs to avoid NAT Gateway cost
# (~$32/month + $0.045/GB). This is the cheapest viable layout for a short-
# lived demo. For production we would use private subnets + NAT or VPC
# endpoints; that decision is captured in ADR 0002 (local-first, cloud-fluent).
# ---------------------------------------------------------------------------
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.8"

  name = "${local.name}-vpc"
  cidr = "10.42.0.0/16"

  azs            = local.azs
  public_subnets = ["10.42.1.0/24", "10.42.2.0/24"]

  enable_dns_hostnames = true
  enable_dns_support   = true

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }
}

# ---------------------------------------------------------------------------
# EKS — smallest viable cluster. SPOT node group. No managed add-ons that
# cost money. The cluster's own control-plane bills at ~$0.10/hour ($2.40/day)
# which is the dominant cost; that's why this stack is designed to be applied
# and destroyed within an hour.
# ---------------------------------------------------------------------------
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.20"

  cluster_name    = local.name
  cluster_version = var.kubernetes_version

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.public_subnets

  cluster_endpoint_public_access = true

  enable_cluster_creator_admin_permissions = true

  eks_managed_node_groups = {
    spot = {
      desired_size   = var.node_desired_size
      min_size       = 1
      max_size       = 3
      instance_types = [var.node_instance_type]
      capacity_type  = var.node_capacity_type
      subnet_ids     = module.vpc.public_subnets
    }
  }
}

# ---------------------------------------------------------------------------
# Cost guardrail. Triggers an email at 80% and 100% of the budget threshold.
# Only created when alert_email is set.
# ---------------------------------------------------------------------------
resource "aws_budgets_budget" "monthly" {
  count = var.alert_email != "" ? 1 : 0

  name              = "${local.name}-monthly"
  budget_type       = "COST"
  limit_amount      = tostring(var.budget_amount_usd)
  limit_unit        = "USD"
  time_unit         = "MONTHLY"
  time_period_start = formatdate("YYYY-MM-DD_hh:mm", timestamp())

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = [var.alert_email]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = [var.alert_email]
  }

  lifecycle {
    ignore_changes = [time_period_start]
  }
}
