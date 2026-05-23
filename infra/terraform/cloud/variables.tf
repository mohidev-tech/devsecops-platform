variable "region" {
  description = "AWS region for the ephemeral cluster."
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "Short cluster name. Suffixed with random_id for uniqueness."
  type        = string
  default     = "devsecops"
}

variable "kubernetes_version" {
  description = "EKS control-plane version."
  type        = string
  default     = "1.30"
}

variable "node_instance_type" {
  description = "Single instance type used for the cheapest viable node group."
  type        = string
  default     = "t3.medium"
}

variable "node_capacity_type" {
  description = "SPOT or ON_DEMAND. SPOT is the cost-discipline default."
  type        = string
  default     = "SPOT"
}

variable "node_desired_size" {
  description = "Node count. Default 2 keeps the platform schedulable; raise only for the demo."
  type        = number
  default     = 2
}

variable "owner" {
  description = "Tag applied to every resource so accidental survivors are easy to find."
  type        = string
  default     = "mohidev-tech"
}

variable "budget_amount_usd" {
  description = "Monthly budget threshold for the alert."
  type        = number
  default     = 5
}

variable "alert_email" {
  description = "Email for the budget alert. Leave empty to skip the budget."
  type        = string
  default     = ""
}
