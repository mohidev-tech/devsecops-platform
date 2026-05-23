# Admission policies

Three artifacts live here, each enforcing a different layer:

| File | Engine | What it enforces |
|---|---|---|
| `kyverno-trusted-registry.yaml` | Kyverno ClusterPolicy | Reject Pods in `app`/`default` whose images aren't from `ghcr.io/mohidev-tech/*` (plus the explicit postgres allowlist) |
| `kyverno-require-resources.yaml` | Kyverno ClusterPolicy | Reject Pods missing CPU/memory requests + limits. Excludes platform namespaces we don't own |
| `disallow-unsigned-images.rego` | OPA/Gatekeeper | Same intent as the Kyverno trusted-registry policy, kept for the Gatekeeper-comparison story |

## Why Kyverno + leave the rego in place?
- Kyverno is the default enforcement path because installing it is one `helm install` and writing rules is YAML (no constraint templates, no rego).
- The OPA rego is kept as a deliberate side-by-side so anyone evaluating the repo can see we know the difference. ADR 0005 (forthcoming) will record the choice.

## Install (after `make cluster && make deploy`)

```bash
helm repo add kyverno https://kyverno.github.io/kyverno/
helm install kyverno kyverno/kyverno -n kyverno --create-namespace --wait

kubectl apply -f security/policies/kyverno-trusted-registry.yaml
kubectl apply -f security/policies/kyverno-require-resources.yaml
```

Watch it work:

```bash
# Should fail with a clear admission error:
kubectl -n app run nginx-test --image=nginx
```

## Phase 3 link
The trusted-registry policy is the *gate* that the [secure-supply-chain](https://github.com/mohidev-tech/secure-supply-chain) repo (satellite 3) feeds: that repo publishes cosign-signed images to `ghcr.io/mohidev-tech/*` with SBOMs attached. The combination — sign at publish, verify at admission — is the SLSA-level story.
