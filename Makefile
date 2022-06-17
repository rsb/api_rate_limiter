SHELL := /bin/bash

VERSION := 1.0
KIND_CLUSTER := limiter-api-cluster
LIMITER_API_DOCKER_IMAGE := limit-api-amd64

tidy:
	go mod tidy && go mod vendor

all: limiter-api

run:
	go run app/cli/limiter/main.go api serve

test:
	go test ./...

limiter-api:
	docker build \
		-f infra/docker/Dockerfile.lola-api \
		-t $(LIMITER_API_DOCKER_IMAGE):$(VERSION) \
		--build-arg VCS_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.



# ==============================================================================
# Running from within k8s/kind


kind-up:
	kind create cluster \
		--image kindest/node:v1.22.0 \
		--name $(KIND_CLUSTER) \
		--config infra/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=limiter-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-restart:
	kubectl rollout restart deployment limiter-pod

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-limiter:
	kubectl get pods -o wide --watch

kind-load:
	cd infra/k8s/kind/limiter-pod; kustomize edit set image limiter-api-image=$(LIMITER_API_DOCKER_IMAGE):$(VERSION)
	kind load docker-image $(LIMITER_API_DOCKER_IMAGE):$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build infra/k8s/kind/database-pod | kubectl apply -f -
	kubectl wait --namespace=database-system --timeout=120s --for=condition=Available deployment/database-pod

	kustomize build infra/k8s/kind/limiter-pod | kubectl apply -f -

kind-describe:
	kubectl describe pod -l app=limiter

kind-logs:
	kubectl logs -l app=limiter --all-containers=true -f --tail=100

kind-service-delete:
	kustomize build infra/k8s/kind/limiter-pod | kubectl delete -f -
	kustomize build infra/k8s/kind/database-pod | kubectl delete -f -

