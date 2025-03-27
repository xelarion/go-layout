# Deployment Guide

## Prerequisites

1. k3s installed on target servers
2. Docker registry access
3. kubectl configured to access the target k3s cluster
4. Git installed
5. yq tool installed for YAML processing

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

1. API Service:
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

### API Service

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
# Scale API service
kubectl scale deployment api-deployment -n go-layout --replicas=<number>

# Scale Task service
kubectl scale deployment task-deployment -n go-layout --replicas=<number>
