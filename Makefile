.PHONY: cluster destroy images deploy smoke argocd vault observe test lint clean

cluster:
	kind create cluster --config deploy/kind/cluster.yaml || true
	kubectl cluster-info --context kind-devsecops

destroy:
	kind delete cluster --name devsecops

images:
	docker build -t ghcr.io/mohidev-tech/devsecops-platform-api:0.1.0 services/api
	docker build -t ghcr.io/mohidev-tech/devsecops-platform-worker:0.1.0 services/worker
	kind load docker-image ghcr.io/mohidev-tech/devsecops-platform-api:0.1.0 --name devsecops
	kind load docker-image ghcr.io/mohidev-tech/devsecops-platform-worker:0.1.0 --name devsecops

deploy: images
	helm upgrade --install app deploy/helm/api  -n app --create-namespace
	helm upgrade --install app deploy/helm/worker -n app

smoke:
	kubectl -n app port-forward svc/app-api 8080:80 &
	sleep 2
	curl -s http://localhost:8080/healthz
	curl -s -X POST http://localhost:8080/api/v1/jobs

argocd:
	kubectl create namespace argocd || true
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
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

clean:
	rm -rf services/api/bin services/worker/bin
