# 部署指南

## 先决条件
1. 目标服务器安装 k3s
2. Docker 镜像仓库访问权限
3. kubectl 配置为访问目标 k3s 集群
4. 安装 Git
5. 安装 yq 工具用于 YAML 处理

## 服务器设置过程

### 1. 配置服务器信息
编辑 `deploy/config/servers.yaml` 文件以添加您要部署到的服务器：
```yaml
servers:
  - name: muggleday
    host: www.muggleday.com
    user: deploy
    k3s_version: v1.28.1+k3s1
    region: cn  # 使用 'cn' 表示中国区域（加速镜像）或 'global' 表示国际区域

  - name: customer1
    host: server1.customer1.com
    user: deploy
    k3s_version: v1.28.1+k3s1
    region: cn  # 使用 'cn' 表示中国区域（加速镜像）或 'global' 表示国际区域
```
`region` 参数决定了用于 k3s 安装的镜像源：
- `cn`：使用中国镜像，在中国大陆安装速度更快
- `global`：使用默认的全球镜像

### 2. 部署到单个服务器
```bash
# 部署到单个服务器
make deploy-server SERVER=muggleday
```

### 3. 部署到所有服务器
```bash
# 部署到配置中的所有服务器
make deploy-all
```

## 部署过程

### 1. 构建和推送镜像
```bash
# 构建和推送镜像
make build
```

### 2. 部署到单节点
```bash
# 部署到单节点 k3s
make deploy-single
```

### 3. 部署到集群
```bash
# 部署到 k3s 集群
make deploy-cluster
```

## 环境配置
所有环境变量都通过 ConfigMaps 和 Secrets 进行管理：
1. ConfigMap (`configmap.yaml`)：
   - 包含非敏感配置
   - 包括数据库、Redis 和 RabbitMQ 连接设置

2. Secret (`secret.yaml`)：
   - 包含敏感信息
   - 包括数据库密码和 RabbitMQ 凭据

## 服务架构
部署包括：
1. Web API 服务：
   - 单节点：1 个副本
   - 集群：3 个副本，具有 pod 反亲和性
   - 通过 Ingress 暴露

2. Task 服务：
   - 单节点：1 个副本
   - 集群：2 个副本，具有 pod 反亲和性

3. 数据库服务：
   - PostgreSQL
   - Redis
   - RabbitMQ

## 资源需求
### Web API 服务
- CPU：200m-500m
- 内存：256Mi-512Mi

### Task 服务
- CPU：100m-300m
- 内存：128Mi-256Mi

### 数据库服务
- PostgreSQL：500m-1000m CPU，512Mi-1Gi 内存，5Gi 存储
- Redis：200m-500m CPU，256Mi-512Mi 内存，1Gi 存储
- RabbitMQ：200m-500m CPU，256Mi-512Mi 内存，1Gi 存储

## 监控和维护
1. 检查 pod 状态：
```bash
kubectl get pods -n go-layout
```

2. 检查服务状态：
```bash
kubectl get svc -n go-layout
```

3. 查看 pod 日志：
```bash
kubectl logs -n go-layout <pod-name>
```

## 故障排除
1. 如果 pod 无法启动：
```bash
kubectl describe pod -n go-layout <pod-name>
```

2. 如果服务无法访问：
```bash
kubectl describe svc -n go-layout <service-name>
```

1. 如果服务器部署失败：
```bash
# 检查 SSH 连接
ssh <user>@<server-host>

# 检查 k3s 安装
ssh <user>@<server-host> "sudo systemctl status k3s"
```

## 安全注意事项
1. 所有敏感数据都存储在 Kubernetes Secrets 中
2. Pod 反亲和性确保集群模式下的高可用性
3. 资源限制防止资源耗尽
4. 容器中使用非 root 用户以增强安全性

## 扩展
要扩展服务：
```bash
# 扩展 Web API 服务
kubectl scale deployment web-api-deployment -n go-layout --replicas=<number>

# 扩展 Task 服务
kubectl scale deployment task-deployment -n go-layout --replicas=<number>
