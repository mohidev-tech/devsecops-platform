# Contributing to devsecops-platform

This is a portfolio reference platform — not a production system. PRs that improve clarity, hardening, or the demo experience are welcome. PRs that bloat scope or add hypothetical features will be politely declined.

## Local development loop

```bash
git clone https://github.com/mohidev-tech/devsecops-platform
cd devsecops-platform

# Bring it up
make secure-mode

# Iterate on a service
cd services/api
go test -race ./...
cd ../..
make images deploy   # rebuild + redeploy
make smoke           # assert end-to-end
```

## Where to make changes

| You want to... | Edit |
|---|---|
| Add a new admission policy | `security/policies/` + reference from `Makefile` `policies` target |
| Add a new Vault-protected secret | `security/vault/policies/*.hcl` + update `security/vault/bootstrap.sh` |
| Add a Prometheus alert | `deploy/helm/api/templates/prometheusrule.yaml` (or a new chart's PrometheusRule) |
| Add a Grafana panel | `observability/grafana/dashboards/slo-api.json` |
| Change the cluster shape | `deploy/kind/cluster.yaml` (local) or `infra/terraform/cloud/main.tf` (EKS) |
| Add a new service | New chart under `deploy/helm/<name>/` + corresponding `deploy/argocd/apps/<name>.yaml` |
| Document a design decision | New file under `docs/adr/` |

## PR checklist

- [ ] `make test lint` passes.
- [ ] `make secure-mode && make smoke` works end-to-end (or you've documented why you couldn't test it).
- [ ] Architecture-level changes have an ADR under `docs/adr/`.
- [ ] If you touched a Helm chart, `helm lint deploy/helm/<chart>` is clean.
- [ ] If you changed Terraform, `terraform fmt` and `terraform validate` are clean for both `infra/terraform/local/` and `infra/terraform/cloud/`.

## Coding conventions

- **Conventional Commits** (`feat:`, `fix:`, `chore:`, `docs:`). The git log is a portfolio asset.
- **One concept per PR.** A PR that adds a Helm chart AND changes the API AND updates docs is three PRs. Reviewers (including future you) appreciate it.
- **No mystery values in YAML.** Every magic number or hardcoded path gets either a `# comment` or a values.yaml entry.
- **ADR-first for architecture changes.** If the change requires explaining "why we did X over Y" to a stranger reading the repo cold, it gets an ADR.

## Hardening contributions are the most valuable

The strongest signal in this repo is the security posture. Examples of welcome PRs:

- Replace dev-mode Vault with an HA Raft install + auto-unseal flow.
- Add a `verifyImages` Kyverno policy importing from [secure-supply-chain](https://github.com/mohidev-tech/secure-supply-chain) and replace the plain trusted-registry check.
- Add a `cspm-scanner` job to CI that scans `infra/terraform/cloud/`.
- Add NetworkPolicy default-deny in the `app` namespace with explicit allow rules.
- Replace the homegrown SLO burn-rate alert with the OpenSLO spec or [Pyrra](https://github.com/pyrra-dev/pyrra).

## License

By submitting a PR you agree the contribution is Apache 2.0 licensed. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
