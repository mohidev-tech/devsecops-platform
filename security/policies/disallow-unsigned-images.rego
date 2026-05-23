package kubernetes.admission

# Reject pods whose images aren't from our signed registry path.
# Pair with cosign verification via Sigstore policy-controller in production.

deny[msg] {
  input.request.kind.kind == "Pod"
  some i
  image := input.request.object.spec.containers[i].image
  not startswith(image, "ghcr.io/mohidev-tech/")
  msg := sprintf("image %q is not from a trusted registry", [image])
}
