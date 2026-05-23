# Vault config

Phase 2 brings Vault up in-cluster, enables K8s auth, defines a policy granting the `api` service account read access to `secret/data/api/db`, and switches the api Deployment to inject DB creds via the Vault agent sidecar.

Files to land in Phase 2:
- `values.yaml` — Helm values for hashicorp/vault chart (dev mode for local; ha for cloud)
- `policies/api.hcl` — read-only policy for app secrets
- `bootstrap.sh` — idempotent setup: enable kv-v2, enable kubernetes auth, write policy, bind role to service account

🚧 Phase 2.
