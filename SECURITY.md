# Security policy

This is a **portfolio reference repo**. It is not deployed anywhere user-facing. The threat model is "a recruiter or interviewer clones it and runs `make secure-mode` on their laptop." Still, security issues matter — the repo's whole point is to demonstrate security engineering.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting:

→ [Report a vulnerability](https://github.com/mohidev-tech/devsecops-platform/security/advisories/new)

Include:

- A short description.
- The component (api / worker / Vault chart / Kyverno policy / Terraform / ...).
- Reproducer steps — ideally a failing `make smoke` or a manifest that bypasses a policy.
- Suggested fix if you have one.

I aim to acknowledge within **72 hours** and fix critical issues within **30 days**.

## Known design caveats — not "vulnerabilities" but worth calling out

These are documented, deliberate trade-offs for portfolio scope. Please don't file reports on them — file PRs that fix them instead.

| Caveat | Why it's like this | Tracked in |
|---|---|---|
| **Vault runs in dev mode** with `root` as the root token | One-pod, in-memory storage; instant bring-up for the demo | [ADR 0004](docs/adr/0004-vault-replaces-plaintext-db-creds.md) |
| **Postgres ships with plaintext dev creds in Phase 1** | Deliberately shown as the *before* state that Vault replaces | [ADR 0003](docs/adr/0003-postgres-job-queue.md), [ADR 0004](docs/adr/0004-vault-replaces-plaintext-db-creds.md) |
| **EKS cluster uses public subnets** with no NAT Gateway | Demo cost discipline — would not be production-appropriate | [infra/terraform/cloud/README.md](infra/terraform/cloud/README.md) |
| **`enable_cluster_creator_admin_permissions = true`** on EKS | Saves 20 minutes of aws-auth troubleshooting on the first apply | inline comment in `main.tf` |
| **Argo CD `selfHeal: true`** | Hostile-PR could be auto-reverted; safer for demo, weaker for real environments | comment in `application.yaml` |

## What's in scope for reports

- Admission policies that can be trivially bypassed.
- Helm chart values that materially weaken pod security (e.g. drops `runAsNonRoot`).
- Workflow secrets exposed in CI logs.
- Path traversal or arbitrary file overwrite via the smoke script or Makefile targets.
- Default Helm values that ship critical or high CVEs in the published images (the api/worker images, not postgres-upstream).

## What's out of scope

- "Vault root token is `root`" — see caveat table.
- "Postgres password is in plaintext" — see caveat table.
- Container CVEs in upstream images we don't build (postgres, kyverno, argocd). Report those to the upstream project.
- Theoretical issues without a concrete reproducer.

## Hardening you can do

If you fork this and run it somewhere real (please don't), the minimum changes:

1. Switch Vault to `ha.enabled=true` with Raft storage + KMS auto-unseal.
2. Delete the `app-postgres` Secret and use Vault dynamic credentials for Postgres instead.
3. Replace EKS public subnets with private subnets + a single NAT (or VPC endpoints for ECR/STS/S3).
4. Enable `enforce_admins=true` on branch protection.
5. Wire [secure-supply-chain](https://github.com/mohidev-tech/secure-supply-chain)'s `verify-signatures` policy to replace the simple `trusted-registry` Kyverno policy.
6. Add NetworkPolicy default-deny in every namespace.

The repo's [ADR 0002](docs/adr/0002-local-first-cloud-fluent.md) explains the local-first/cloud-fluent design — none of these gaps are accidents.
