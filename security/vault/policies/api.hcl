# Read-only access to the app DB secret. Bound to the `api` service account
# in the `app` namespace via kubernetes auth (see bootstrap.sh).
path "secret/data/app/db" {
  capabilities = ["read"]
}
