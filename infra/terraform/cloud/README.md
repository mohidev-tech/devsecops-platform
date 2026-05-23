# Cloud variant — apply briefly, then destroy

This directory provisions the platform in a real cloud (target: AWS EKS) for the "I can do this in the cloud too" demo. **It is designed to be applied, screenshotted, and torn down within an hour.**

## Cost discipline

- Single-AZ EKS, smallest node group, spot instances.
- No NAT gateway (private nodes use VPC endpoints for ECR/S3/STS only).
- All resources tagged `lifecycle=ephemeral` and `owner=mohidev-tech`.
- A budget alert is provisioned at $5/month.

Run the demo:

```bash
terraform init
terraform apply -auto-approve
# ... take screenshots of dashboards, capture cluster logs ...
terraform destroy -auto-approve
```

Verify post-destroy:

```bash
aws ec2 describe-instances --filters "Name=tag:lifecycle,Values=ephemeral" \
  --query "Reservations[].Instances[?State.Name!='terminated']"
# must return []
```

## Status

🚧 Stub — implement in Phase 1 once local platform is green.
