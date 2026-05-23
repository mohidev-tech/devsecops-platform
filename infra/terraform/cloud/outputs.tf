output "cluster_name" {
  value = module.eks.cluster_name
}

output "region" {
  value = var.region
}

output "kubeconfig_command" {
  description = "Run this to point kubectl at the cluster."
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}

output "destroy_reminder" {
  description = "Run this when the demo is done. Control-plane bills hourly."
  value       = "terraform destroy -auto-approve"
}
