# ADR 0002 — Local-first, cloud-fluent

## Status
Accepted

## Context
A portfolio platform must demonstrate cloud competence without running up a bill. Running a continuously-on EKS cluster across the build period would cost hundreds of dollars and offers no signal beyond what an ephemeral demo provides.

## Decision
Default execution target is **local** (`kind` provisioned by Terraform). A second Terraform root, `infra/terraform/cloud/`, provisions an equivalent **EKS** cluster designed to be applied, screenshotted, and destroyed within an hour. Both share Helm charts, Argo CD config, and observability stack — only the cluster provider differs.

## Consequences
- ✅ Reproducible $0 demo for reviewers without cloud access.
- ✅ Genuine cloud proof exists when needed.
- ✅ Cost story headlines the cloud README — reads as senior judgment.
- ⚠️ Cloud-only services (e.g., AWS Secrets Manager, IAM Roles for Service Accounts) need explicit local equivalents (Vault, K8s ServiceAccount tokens). Documented as we go.
