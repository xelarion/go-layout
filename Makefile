# Project version (use git tag)
VERSION := $(shell git describe --tags --always --dirty)

# Docker registry
REGISTRY ?= registry.example.com

# Service names
SERVICES = api task

# Build targets
.PHONY: build build-api build-task

build: build-api build-task
	@echo "Build complete"

build-api:
	docker build \
		--build-arg SERVICE=api \
		--build-arg CONFIG_ENV=prod \
		-t ${REGISTRY}/go-layout-api:${VERSION} .

docker push ${REGISTRY}/go-layout-api:${VERSION}

build-task:
	docker build \
		--build-arg SERVICE=task \
		--build-arg CONFIG_ENV=prod \
		-t ${REGISTRY}/go-layout-task:${VERSION} .

docker push ${REGISTRY}/go-layout-task:${VERSION}

# Deploy targets
.PHONY: deploy deploy-single deploy-cluster deploy-server deploy-all

deploy: deploy-single

deploy-single:
	@echo "Deploying to single node k3s..."
	kubectl apply -f deploy/k3s/single/namespace.yaml
	kubectl apply -f deploy/k3s/single/configmap.yaml
	kubectl apply -f deploy/k3s/single/secret.yaml
	kubectl apply -f deploy/k3s/single/database.yaml
	kubectl apply -f deploy/k3s/single/redis.yaml
	kubectl apply -f deploy/k3s/single/rabbitmq.yaml
	kubectl apply -f deploy/k3s/single/api-deployment.yaml
	kubectl apply -f deploy/k3s/single/task-deployment.yaml
	kubectl apply -f deploy/k3s/single/ingress.yaml

deploy-cluster:
	@echo "Deploying to k3s cluster..."
	kubectl apply -f deploy/k3s/cluster/namespace.yaml
	kubectl apply -f deploy/k3s/cluster/configmap.yaml
	kubectl apply -f deploy/k3s/cluster/secret.yaml
	kubectl apply -f deploy/k3s/cluster/database.yaml
	kubectl apply -f deploy/k3s/cluster/redis.yaml
	kubectl apply -f deploy/k3s/cluster/rabbitmq.yaml
	kubectl apply -f deploy/k3s/cluster/api-deployment.yaml
	kubectl apply -f deploy/k3s/cluster/task-deployment.yaml
	kubectl apply -f deploy/k3s/cluster/ingress.yaml

deploy-server:
	@echo "Deploying to server: $(SERVER)"
	./deploy/scripts/deploy-server.sh $(SERVER)

deploy-all:
	@echo "Deploying to all servers..."
	./deploy/scripts/deploy-all.sh

# Clean targets
.PHONY: clean

clean:
	docker rmi ${REGISTRY}/go-layout-api:${VERSION} || true
	docker rmi ${REGISTRY}/go-layout-task:${VERSION} || true

# Help target
.PHONY: help

help:
	@echo "Makefile Usage:"
	@echo "  make build           - Build and push all Docker images"
	@echo "  make deploy          - Deploy to single node k3s (default)"
	@echo "  make deploy-single   - Deploy to single node k3s"
	@echo "  make deploy-cluster  - Deploy to k3s cluster"
	@echo "  make deploy-server SERVER=muggleday  - Deploy to specific server"
	@echo "  make deploy-all      - Deploy to all servers"
	@echo "  make clean           - Clean up Docker images"
	@echo "  make help            - Show this help message"