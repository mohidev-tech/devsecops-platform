.PHONY: cluster destroy images deploy postgres api worker smoke argocd vault observe test lint clean

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

argocd:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
	kubectl -n argocd rollout status deploy/argocd-server --timeout=180s
	kubectl apply -f deploy/argocd/application.yaml

vault:
	@echo "Phase 2 — see security/vault/README.md"

observe:
	@echo "Phase 2 — Prometheus + Grafana bring-up"

test:
	cd services/api    && go test -race -count=1 ./...
	cd services/worker && go test -race -count=1 ./...

lint:
	cd services/api    && go vet ./...
	cd services/worker && go vet ./...
	helm lint deploy/helm/postgres deploy/helm/api deploy/helm/worker

clean:
	rm -rf services/api/bin services/worker/bin
