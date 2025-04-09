# Project version (use git tag)
VERSION := $(shell git describe --tags --always --dirty)

# Docker registry
REGISTRY ?= docker.io

# Service names
SERVICES = web-api task

# Database migration commands
.PHONY: migrate migrate-up migrate-down migrate-status migrate-create migrate-reset
.PHONY: migrate-version migrate-redo migrate-up-to migrate-down-to migrate-fix migrate-ci

# Run database migrations (alias for migrate-up)
migrate: migrate-up

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

.PHONY: gen-models

# Generate models from database schema
gen-models:
	@echo "Generating models from database schema..."
	@mkdir -p internal/model/gen
	@go run tools/gen/main.go

.PHONY: swagger-install swagger-gen swagger-web swagger-autocomment swagger-all

# Swagger commands
# Install Swagger tools
swagger-install:
	@echo "Installing Swagger tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go get -u github.com/swaggo/gin-swagger
	go get -u github.com/swaggo/files

# Generate Swagger documentation for Web API
swagger-web:
	@echo "Generating Swagger documentation for Web API..."
	cd internal/api/http/web && swag init -g swagger/doc.go -o swagger/docs && swag fmt

# Generate intelligent Swagger comments for handler methods
swagger-autocomment:
	@echo "Generating intelligent Swagger comments for handler methods..."
	go run tools/swagger_autocomment/main.go -dir ./internal/api/http/web/handler

# Generate all Swagger documentation (comments and docs)
swagger-all: swagger-autocomment swagger-web
	@echo "All Swagger documentation generated successfully"

# Generate Swagger documentation for all APIs
swagger-gen: swagger-web
	@echo "Swagger documentation generated successfully"


# Build targets
.PHONY: build build-web-api build-task build-migrate

# Build and push all Docker images
build: build-web-api build-task build-migrate
	@echo "Build complete"

# Build and push Web API Docker image
build-web-api:
	docker build \
		--build-arg SERVICE=web-api \
		--build-arg CONFIG_ENV=prod \
		-t ${REGISTRY}/go-layout-web-api:${VERSION} .

	docker push ${REGISTRY}/go-layout-web-api:${VERSION}

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
	kubectl apply -f deploy/k3s/single/web-api-deployment.yaml
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
	kubectl apply -f deploy/k3s/cluster/web-api-deployment.yaml
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
	docker rmi ${REGISTRY}/go-layout-web-api:${VERSION} || true
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
