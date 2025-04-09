# Deployment Guide

This document outlines the process for deploying the go-layout application to production environments.

## Overview

The go-layout application is designed for containerized deployment using Docker and Kubernetes (specifically k3s). The application consists of:

- **Web API Service**: The main web API service
- **Task Service**: Background task processing service
- **Database Migrations**: Separate container for running database migrations

## Prerequisites

- Docker and Docker Compose for local development
- kubectl for interacting with Kubernetes clusters
- Access to a Docker registry (default: docker.io)
- A running k3s cluster (single node or multi-node)

## Building Docker Images

All services are packaged as Docker containers. The following commands build and push the images:

```bash
# Build and push all images (Web API, Task, Migrations)
make build

# Build and push individual services
make build-web-api
make build-task
make build-migrate
```

The version tag for the images is automatically determined from git tags/commits.

## Database Migrations

Before deploying the application, database migrations should be run to ensure the database schema is up to date.

### Running Migrations for Production

For production deployments, migrations are run via a Kubernetes job:

```bash
# Run migrations on a single-node k3s
make deploy-migrate

# Run migrations on a multi-node k3s cluster
make deploy-migrate-cluster
```

## Deployment Options

### Single-Node Deployment

For simple deployments on a single k3s node:

```bash
make deploy-single
```

This will deploy:
- The Web API service
- The Task service
- Supporting services (if configured in the k3s manifests)
- Ingress configuration for external access

### Cluster Deployment

For deploying to a multi-node k3s cluster:

```bash
make deploy-cluster
```

This deploys the same components but with appropriate configuration for a distributed environment.

### Server-Specific Deployment

To deploy to a specific server:

```bash
make deploy-server SERVER=<server_name>
```

### Multi-Server Deployment

To deploy to all servers:

```bash
make deploy-all
```

## Configuration

### Environment Configuration

Production configuration is stored in `config/prod/.env` and is packaged into the Docker images during build.

Important environment variables for production:

- `GO_ENV=prod`: Sets the application to run in production mode
- `HTTP_MODE=release`: Configures Gin to run in release mode
- `PG_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string (if Redis is used)
- `RABBITMQ_URL`: RabbitMQ connection string (if RabbitMQ is used)
- `JWT_SECRET`: Secret key for JWT authentication
- `HTTP_ALLOW_ORIGINS`: CORS allowed origins

### Kubernetes Configuration

Kubernetes-specific configuration is stored in:
- `deploy/k3s/single/` for single-node deployments
- `deploy/k3s/cluster/` for multi-node cluster deployments

These include:
- ConfigMaps and Secrets for application configuration
- Deployment manifests for services
- Service definitions
- Ingress configuration
- Resource limits and requests

## Health Monitoring

The application provides health check endpoints that are used by Kubernetes to monitor service health:
- Liveness probes check if the service is running
- Readiness probes check if the service is able to handle requests

## Scaling

The application is designed to be stateless and can be scaled horizontally by increasing the number of replicas:

```bash
kubectl scale deployment/web-api-deployment --replicas=3 -n go-layout
kubectl scale deployment/task-deployment --replicas=2 -n go-layout
```

## Rollback Procedure

To roll back to a previous version:

1. Identify the previous version tag
2. Update the deployment with the previous version:

```bash
kubectl set image deployment/web-api-deployment web-api=${REGISTRY}/go-layout-web-api:${PREVIOUS_VERSION} -n go-layout
kubectl set image deployment/task-deployment task=${REGISTRY}/go-layout-task:${PREVIOUS_VERSION} -n go-layout
```

## Troubleshooting

### Checking Logs

```bash
# View Web API service logs
kubectl logs -l app=web-api -n go-layout

# View Task service logs
kubectl logs -l app=task -n go-layout

# View migration job logs
kubectl logs -l app=db-migrate -n go-layout
```

### Checking Pod Status

```bash
kubectl get pods -n go-layout
```

### Common Issues

1. **Database Connection Failures**: Check PG_URL environment variable and network connectivity
2. **Memory/CPU Limits**: Check if pods are hitting resource limits
3. **Ingress Issues**: Verify ingress configuration and TLS certificates

## Server Setup Process

### 1. Configure Server Information

Edit the `deploy/config/servers.yaml` file to add the servers you want to deploy to:

```yaml
servers:
  - name: muggleday
    host: www.muggleday.com
    user: deploy
    k3s_version: v1.28.1+k3s1
    region: cn    # Use 'cn' for China region (faster mirror) or 'global' for international

  - name: customer1
    host: server1.customer1.com
    user: deploy
    k3s_version: v1.28.1+k3s1
    region: cn    # Use 'cn' for China region (faster mirror) or 'global' for international
```

The `region` parameter determines which mirror to use for k3s installation:

- `cn`: Uses the China mirror for faster installation in mainland China
- `global`: Uses the default global mirror

### 2. Deploy to a Single Server

```bash
# Deploy to a single server
make deploy-server SERVER=muggleday
```

### 3. Deploy to All Servers

```bash
# Deploy to all servers in the configuration
make deploy-all
```

## Deployment Process

### 1. Build and Push Images

```bash
# Build and push images
make build
```

### 2. Deploy to Single Node

```bash
# Deploy to single node k3s
make deploy-single
```

### 3. Deploy to Cluster

```bash
# Deploy to k3s cluster
make deploy-cluster
```

## Environment Configuration

All environment variables are managed through ConfigMaps and Secrets:

1. ConfigMap (`configmap.yaml`):
   - Contains non-sensitive configuration
   - Includes database, Redis, and RabbitMQ connection settings

2. Secret (`secret.yaml`):
   - Contains sensitive information
   - Includes database passwords and RabbitMQ credentials

## Service Architecture

The deployment consists of:

1. Web API Service:
   - Single node: 1 replica
   - Cluster: 3 replicas with pod anti-affinity
   - Exposed via Ingress

2. Task Service:
   - Single node: 1 replica
   - Cluster: 2 replicas with pod anti-affinity

3. Database Services:
   - PostgreSQL
   - Redis
   - RabbitMQ

## Resource Requirements

### Web API Service

- CPU: 200m-500m
- Memory: 256Mi-512Mi

### Task Service

- CPU: 100m-300m
- Memory: 128Mi-256Mi

### Database Services

- PostgreSQL: 500m-1000m CPU, 512Mi-1Gi Memory, 5Gi Storage
- Redis: 200m-500m CPU, 256Mi-512Mi Memory, 1Gi Storage
- RabbitMQ: 200m-500m CPU, 256Mi-512Mi Memory, 1Gi Storage

## Monitoring and Maintenance

1. Check pod status:

```bash
kubectl get pods -n go-layout
```

2. Check service status:

```bash
kubectl get svc -n go-layout
```

3. View pod logs:

```bash
kubectl logs -n go-layout <pod-name>
```

## Troubleshooting

1. If pods are not starting:

```bash
kubectl describe pod -n go-layout <pod-name>
```

2. If services are not accessible:

```bash
kubectl describe svc -n go-layout <service-name>
```

1. If server deployment fails:

```bash
# Check the SSH connection
ssh <user>@<server-host>

# Check the k3s installation
ssh <user>@<server-host> "sudo systemctl status k3s"
```

## Security Considerations

1. All sensitive data is stored in Kubernetes Secrets
2. Pod anti-affinity ensures high availability in cluster mode
3. Resource limits prevent resource exhaustion
4. Non-root user is used in containers for security

## Scaling

To scale services:

```bash
# Scale Web API service
kubectl scale deployment api-deployment -n go-layout --replicas=<number>

# Scale Task service
kubectl scale deployment task-deployment -n go-layout --replicas=<number>
