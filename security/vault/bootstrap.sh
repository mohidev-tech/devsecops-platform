#!/usr/bin/env bash
# Idempotent Vault setup. Run after `helm install vault hashicorp/vault -f values.yaml`.
#
# What this does:
#   1. Enable kv-v2 at secret/
#   2. Write the DB credentials to secret/data/app/db
#   3. Write the api read-only policy
#   4. Enable kubernetes auth
#   5. Bind the api ServiceAccount in `app` namespace to the api policy
#
# After this, the api Deployment with vault-agent annotations will get
# DATABASE_URL injected at /vault/secrets/database-url. The plaintext
# Secret from the postgres chart is no longer the source of truth for
# the app — see ADR 0004.

set -euo pipefail

NS="${NS:-vault}"
APP_NS="${APP_NS:-app}"
ROOT_TOKEN="${VAULT_ROOT_TOKEN:-root}"

vex() { kubectl -n "$NS" exec -i vault-0 -- /bin/sh -c "VAULT_TOKEN=$ROOT_TOKEN $*"; }

echo "[vault] enabling kv-v2 at secret/"
vex 'vault secrets enable -path=secret kv-v2' 2>/dev/null || echo "  already enabled"

echo "[vault] writing secret/data/app/db"
vex "vault kv put secret/app/db url='postgres://app:dev-only-replace-in-phase-2@app-postgres.${APP_NS}.svc:5432/jobs?sslmode=disable'"

echo "[vault] writing api policy"
kubectl -n "$NS" cp security/vault/policies/api.hcl vault-0:/tmp/api.hcl
vex 'vault policy write api /tmp/api.hcl'

echo "[vault] enabling kubernetes auth"
vex 'vault auth enable kubernetes' 2>/dev/null || echo "  already enabled"

echo "[vault] configuring kubernetes auth"
vex 'vault write auth/kubernetes/config \
  kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443"'

echo "[vault] binding api ServiceAccount"
vex "vault write auth/kubernetes/role/api \
  bound_service_account_names=api \
  bound_service_account_namespaces=${APP_NS} \
  policies=api \
  ttl=1h"

echo "[vault] done"
