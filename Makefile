# Project version (use git tag)
VERSION := $(shell git describe --tags --always --dirty)

# Docker registry
REGISTRY ?= docker.io

# Service names
SERVICES = api task

# Database migration commands
.PHONY: db migrate migrate-up migrate-down migrate-status migrate-create migrate-reset migrate-version migrate-redo migrate-up-to migrate-down-to migrate-fix migrate-ci

# Run database migrations (simple alias for migrate-up)
db:
	go run cmd/migrate/main.go up

# Apply all pending migrations
migrate-up:
	go run cmd/migrate/main.go up

# Rollback the last migration
migrate-down:
	go run cmd/migrate/main.go down

# Print migration status
migrate-status:
	go run cmd/migrate/main.go status

# Rollback all migrations
migrate-reset:
	go run cmd/migrate/main.go reset

# Create a new migration file (requires NAME=<name>)
migrate-create:
	@[ "${NAME}" ] || ( echo "Please provide NAME=<migration_name>"; exit 1 )
	go run cmd/migrate/main.go create ${NAME} sql

# Print current migration version
migrate-version:
	go run cmd/migrate/main.go version

# Rollback and reapply the latest migration
migrate-redo:
	go run cmd/migrate/main.go redo

# Migrate up to a specific version (requires VERSION=<num>)
migrate-up-to:
	@[ "${VERSION}" ] || ( echo "Please provide VERSION=<version_number>"; exit 1 )
	go run cmd/migrate/main.go up-to ${VERSION}

# Migrate down to a specific version (requires VERSION=<num>)
migrate-down-to:
	@[ "${VERSION}" ] || ( echo "Please provide VERSION=<version_number>"; exit 1 )
	go run cmd/migrate/main.go down-to ${VERSION}

# Fix migrations order (convert timestamps to sequential for production)
migrate-fix:
	go run cmd/migrate/main.go fix

# Prepare migrations for CI/production by fixing versioning
migrate-ci:
	@echo "Preparing migrations for production deployment..."
	go run cmd/migrate/main.go fix

# Run database migrations (alias for migrate-up)
migrate: migrate-up

# Build targets
.PHONY: build build-api build-task build-migrate

# Build and push all Docker images
build: build-api build-task build-migrate
	@echo "Build complete"

# Build and push API Docker image
build-api:
	docker build \
		--build-arg SERVICE=api \
		--build-arg CONFIG_ENV=prod \
		-t ${REGISTRY}/go-layout-api:${VERSION} .

	docker push ${REGISTRY}/go-layout-api:${VERSION}

# Build and push Task Docker image
build-task:
	docker build \
		--build-arg SERVICE=task \
		--build-arg CONFIG_ENV=prod \
		-t ${REGISTRY}/go-layout-task:${VERSION} .

	docker push ${REGISTRY}/go-layout-task:${VERSION}

# Build and push Migration Docker image
build-migrate:
	@echo "Preparing migrations for production deployment..."
	go run cmd/migrate/main.go fix
	docker build \
		-f Dockerfile.migrate \
		--build-arg CONFIG_ENV=prod \
		-t ${REGISTRY}/go-layout-migrate:${VERSION} .

	docker push ${REGISTRY}/go-layout-migrate:${VERSION}

# Deploy targets
.PHONY: deploy deploy-single deploy-cluster deploy-server deploy-all deploy-with-preload preload-images deploy-migrate deploy-migrate-cluster

# Deploy to single node k3s (default)
deploy: deploy-single

# Deploy to single node k3s
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

# Deploy to k3s cluster
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

# Run migrations on single node k3s
deploy-migrate:
	@echo "Running database migrations..."
	cat deploy/k3s/single/migrate-job.yaml | \
		sed "s|\${REGISTRY}|${REGISTRY}|g" | \
		sed "s|\${VERSION}|${VERSION}|g" | \
		kubectl apply -f -

# Run migrations on k3s cluster
deploy-migrate-cluster:
	@echo "Running database migrations in cluster..."
	cat deploy/k3s/cluster/migrate-job.yaml | \
		sed "s|\${REGISTRY}|${REGISTRY}|g" | \
		sed "s|\${VERSION}|${VERSION}|g" | \
		kubectl apply -f -

# Deploy to specific server (requires SERVER=<server_name>)
deploy-server:
	@[ "${SERVER}" ] || ( echo "Please provide SERVER=<server_name>"; exit 1 )
	@echo "Deploying to server: ${SERVER}"
	./deploy/scripts/deploy-server.sh ${SERVER}

# Preload essential K3s images on the server (requires SERVER=<server_name>)
preload-images:
	@[ "${SERVER}" ] || ( echo "Please provide SERVER=<server_name>"; exit 1 )
	@echo "Preloading images for server: ${SERVER}"
	./deploy/scripts/preload-k3s-images.sh ${SERVER}

# Deploy to specific server with preloaded images (requires SERVER=<server_name>)
deploy-with-preload:
	@[ "${SERVER}" ] || ( echo "Please provide SERVER=<server_name>"; exit 1 )
	@echo "Preloading images and deploying to server: ${SERVER}"
	./deploy/scripts/preload-k3s-images.sh ${SERVER}
	./deploy/scripts/deploy-server.sh ${SERVER}

# Deploy to all servers
deploy-all:
	@echo "Deploying to all servers..."
	./deploy/scripts/deploy-all.sh

# Clean targets
.PHONY: clean

# Clean up Docker images
clean:
	docker rmi ${REGISTRY}/go-layout-api:${VERSION} || true
	docker rmi ${REGISTRY}/go-layout-task:${VERSION} || true
	docker rmi ${REGISTRY}/go-layout-migrate:${VERSION} || true

# Show this help message
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
