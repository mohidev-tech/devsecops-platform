# Argo CD — GitOps delivery

## Bootstrap

```bash
make argocd      # installs Argo CD, applies the AppProject + root app
```

After that, **`git push` is the deploy mechanism.** Merging to `main` triggers Argo CD to reconcile the live cluster against `deploy/`.

## Layout

```
application.yaml      AppProject + the single root "app-of-apps" Application
apps/                 Child Applications managed by the root app
  postgres.yaml       sync-wave: -10  (DB comes up first)
  api.yaml            sync-wave: 0
  worker.yaml         sync-wave: 0
```

`sync-wave` annotations sequence the bring-up so api/worker don't crashloop waiting on the database. The same pattern adds Vault + Kyverno + the observability stack in their own waves once Phase 2 lands those as Apps too.

## Verify

```bash
kubectl -n argocd get applications
# NAME       SYNC STATUS   HEALTH STATUS
# root       Synced        Healthy
# postgres   Synced        Healthy
# api        Synced        Healthy
# worker    Synced        Healthy

kubectl -n argocd port-forward svc/argocd-server 8081:443
# open https://localhost:8081 — admin password:
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d
```

## Why app-of-apps?
Adding a new service (or moving Vault/Kyverno/Prometheus under Argo CD) becomes a single PR adding one YAML under `apps/` — no Argo CD config changes needed. The root app picks it up on the next sync.
