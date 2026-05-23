# ADR 0004 — Vault sidecar injection replaces plaintext DB creds

## Status
Accepted

## Context
ADR 0003 acknowledged that Phase 1 ships plaintext Postgres credentials in `deploy/helm/postgres/values.yaml`. They're rendered into a Kubernetes Secret and consumed by api and worker via `secretKeyRef`. This works, but every reader with `get secrets` in the `app` namespace can decode the credentials, and they sit unrotated in git history.

Phase 2 needs to fix this without making the local dev flow painful.

## Decision
Add Vault to the cluster in dev mode (single pod, in-memory storage, root token = `root`). Use the **vault-agent injector** (sidecar pattern) to render the DATABASE_URL into the api pod's filesystem at `/vault/secrets/database-url`. The api process reads this path via a new `DATABASE_URL_FILE` env var, falling back to the plaintext `DATABASE_URL` env when Vault is disabled.

Authentication uses Vault's **kubernetes auth method**: the api ServiceAccount's projected JWT is exchanged for a Vault token bound to a read-only policy (`api.hcl`) that grants read on `secret/data/app/db` and nothing else.

Toggle: `--set vault.enabled=true` on the api Helm chart. Off by default so Phase 1 stays trivially reproducible.

## Why a file, not a shell-sourced env
The api image is `distroless/static:nonroot` — no shell, no `source` command. The common Vault tutorial pattern of `["/bin/sh", "-c", "source /vault/secrets/x && exec /app"]` would force us off distroless and reintroduce a CVE-rich base image. Reading the path directly in Go keeps the hardened image intact.

Side benefit: secret rotation via the vault-agent template no longer requires a pod restart — the api re-reads the file on next connection attempt if we wire pgxpool to reconnect on auth failure (Phase 2.1 stretch).

## Consequences
- ✅ DB creds no longer in `kubectl get secret -o yaml` output.
- ✅ The audit trail of who-read-the-secret is in Vault audit logs, not kube audit logs.
- ✅ The exact same diff that adds this also documents the policy boundary — `api.hcl` is the spec of what the api is *allowed* to know.
- ⚠️ Dev mode Vault stores everything in memory; pod restart wipes the secret. The bootstrap script is idempotent so it's a one-command recovery. Production Vault would use Raft storage with auto-unseal.
- ⚠️ Vault dev mode root token is `root` and is in `values.yaml`. This is fine because dev mode is explicitly not for prod; ADR 0005 will record the production HA setup when we do it.
- 📝 The plaintext path is still in the chart so reviewers can see the *before*. We're not deleting Phase 1 — we're showing the upgrade.
