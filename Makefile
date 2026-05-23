.PHONY: cluster destroy images deploy postgres api worker smoke \
        argocd vault vault-bootstrap policies observe observe-dashboards \
        secure-mode test lint clean

REGISTRY ?= ghcr.io/mohidev-tech
TAG      ?= 0.1.0

cluster:
	kind create cluster --config deploy/kind/cluster.yaml || true
	kubectl cluster-info --context kind-devsecops

destroy:
	kind delete cluster --name devsecops

images:
	docker build -t $(REGISTRY)/devsecops-platform-api:$(TAG)    services/api
	docker build -t $(REGISTRY)/devsecops-platform-worker:$(TAG) services/worker
	kind load docker-image $(REGISTRY)/devsecops-platform-api:$(TAG)    --name devsecops
	kind load docker-image $(REGISTRY)/devsecops-platform-worker:$(TAG) --name devsecops

postgres:
	kubectl create namespace app --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install app deploy/helm/postgres -n app --wait --timeout 2m

api:
	helm upgrade --install app deploy/helm/api -n app --wait --timeout 2m \
		--set image.repository=$(REGISTRY)/devsecops-platform-api \
		--set image.tag=$(TAG)

worker:
	helm upgrade --install app deploy/helm/worker -n app --wait --timeout 2m \
		--set image.repository=$(REGISTRY)/devsecops-platform-worker \
		--set image.tag=$(TAG)

deploy: images postgres api worker

smoke:
	bash scripts/smoke.sh

# -- Phase 2 ---------------------------------------------------------------

argocd:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
	kubectl -n argocd rollout status deploy/argocd-server --timeout=180s
	kubectl apply -f deploy/argocd/application.yaml

vault:
	helm repo add hashicorp https://helm.releases.hashicorp.com 2>/dev/null || true
	helm repo update
	kubectl create namespace vault --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install vault hashicorp/vault -n vault -f security/vault/values.yaml --wait --timeout 3m

vault-bootstrap: vault
	bash security/vault/bootstrap.sh
	# Switch the api to vault-injected creds. Pod restarts pick up the new template.
	helm upgrade --install app deploy/helm/api -n app --wait --timeout 2m \
		--set image.repository=$(REGISTRY)/devsecops-platform-api \
		--set image.tag=$(TAG) \
		--set vault.enabled=true

policies:
	helm repo add kyverno https://kyverno.github.io/kyverno/ 2>/dev/null || true
	helm repo update
	kubectl create namespace kyverno --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install kyverno kyverno/kyverno -n kyverno --wait --timeout 3m
	kubectl apply -f security/policies/kyverno-trusted-registry.yaml
	kubectl apply -f security/policies/kyverno-require-resources.yaml

observe:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
	helm repo update
	kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install kps prometheus-community/kube-prometheus-stack \
		-n monitoring -f observability/prometheus/values.yaml --wait --timeout 5m
	# Re-deploy api so the ServiceMonitor + PrometheusRule land
	helm upgrade --install app deploy/helm/api -n app --wait --timeout 2m \
		--set image.repository=$(REGISTRY)/devsecops-platform-api \
		--set image.tag=$(TAG) \
		--set serviceMonitor.enabled=true \
		--set prometheusRule.enabled=true \
		--set vault.enabled=true
	$(MAKE) observe-dashboards

observe-dashboards:
	kubectl -n monitoring create configmap api-slo-dashboard \
		--from-file=slo-api.json=observability/grafana/dashboards/slo-api.json \
		--dry-run=client -o yaml | kubectl apply -f -
	kubectl -n monitoring label configmap api-slo-dashboard grafana_dashboard=1 --overwrite

# Full secure-mode bring-up — what you'd run for a demo recording.
secure-mode: cluster deploy policies vault-bootstrap observe smoke
	@echo ""
	@echo "Platform is up in secure mode."
	@echo "  Argo CD:  make argocd; kubectl -n argocd port-forward svc/argocd-server 8081:443"
	@echo "  Vault UI: kubectl -n vault port-forward svc/vault 8200:8200   (token: root)"
	@echo "  Grafana:  kubectl -n monitoring port-forward svc/kps-grafana 3000:80  (admin/admin)"

# -- dev ------------------------------------------------------------------

test:
	cd services/api    && go test -race -count=1 ./...
	cd services/worker && go test -race -count=1 ./...

lint:
	cd services/api    && go vet ./...
	cd services/worker && go vet ./...
	helm lint deploy/helm/postgres deploy/helm/api deploy/helm/worker

clean:
	rm -rf services/api/bin services/worker/bin
