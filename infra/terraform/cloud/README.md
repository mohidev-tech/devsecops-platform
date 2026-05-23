# Cloud variant — apply briefly, then destroy

This directory provisions the platform on **AWS EKS** for the "I can do this in the cloud too" demo. **It is designed to be applied, screenshotted, and torn down within an hour.**

## Cost discipline (what you pay for)

| Resource | Cost driver | Mitigation |
|---|---|---|
| **EKS control plane** | `$0.10/hour` flat regardless of usage | Apply → demo → destroy within an hour |
| **EC2 nodes** | `t3.medium` × 2 spot | ~$0.013/hour each, ~$0.03/hour total |
| **NAT Gateway** | `~$0.045/hour` + per-GB | **Removed.** Nodes use public subnets + public IPs |
| **EBS root volumes** | $0.10/GB-month | 20GB × 2, prorated to ~cents for an hour |
| **Data transfer** | `$0.09/GB` egress | Minimal — `kubectl` traffic only |

**Realistic 1-hour demo cost: ~$0.15.** Budget alert at $5 catches a forgotten cluster within a day.

## Run the demo

```bash
cd infra/terraform/cloud

# Optional: enable budget alert
export TF_VAR_alert_email="you@example.com"

terraform init
terraform apply -auto-approve     # ~10 minutes
$(terraform output -raw kubeconfig_command)

# Verify
kubectl get nodes
kubectl get ns

# Deploy the platform on the cloud cluster
cd ../../..
make deploy
make smoke

# Take screenshots of grafana, argocd, vault — then:
cd infra/terraform/cloud
terraform destroy -auto-approve   # ~5 minutes
```

## Verify nothing survived

```bash
aws ec2 describe-instances \
  --filters "Name=tag:project,Values=devsecops-platform" \
  --query "Reservations[].Instances[?State.Name!='terminated'].InstanceId"
# Must return []

aws eks list-clusters | grep devsecops || echo "clean"
```

## Design choices

- **Public subnets only.** Skips the NAT Gateway, which is the single biggest cost trap for short-lived EKS demos. For a real workload we'd use private subnets + NAT or VPC endpoints — but that quadruples the demo cost.
- **SPOT instances.** ~70% cheaper than on-demand. Eviction is fine for a demo; if it happens, you get to demonstrate the platform recovering, which is itself a portfolio talking point.
- **Single node group, 2 nodes.** Just enough to run postgres + api×2 + worker + observability without scheduling pressure.
- **`enable_cluster_creator_admin_permissions = true`.** The TF runner gets cluster-admin automatically — saves 20 minutes of IAM/aws-auth troubleshooting on first apply.
