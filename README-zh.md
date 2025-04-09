# Go Layout

一个使用Go构建的可扩展、高性能、高可用性的Web应用模板。

## 特点

- **分层架构**：采用Handler → Service → Usecase → Repository分层设计模式，实现关注点明确分离
- **高性能**：优化的数据库查询、高效的缓存策略和性能监控
- **高可用性**：无状态设计，便于水平扩展和集群部署
- **可扩展性**：模块化设计，支持水平和垂直扩展
- **整洁代码**：结构良好、可维护、可扩展的代码库
- **安全性**：JWT认证、输入验证和防范常见漏洞的保护
- **错误处理**：全面的错误处理系统，包含结构化日志和自定义错误类型
- **中间件**：模块化的中间件组件，用于请求处理、错误处理、请求超时控制和CORS
- **任务管理**：集成支持定时任务、轮询任务和基于队列的异步处理

## 技术栈

- **编程语言**：Go
- **Web框架**：[Gin](https://github.com/gin-gonic/gin)
- **ORM**：[GORM](https://gorm.io/)
- **数据库**：PostgreSQL（可配置切换至MySQL）
- **缓存**：Redis（可插拔，如不需要可移除）
- **消息队列**：RabbitMQ（可插拔，如不需要可移除）
- **日志**：使用[zap](https://github.com/uber-go/zap)进行结构化日志记录
- **配置**：基于环境的配置，使用[github.com/caarlos0/env/v11](https://github.com/caarlos0/env)
- **错误处理**：自定义错误包，支持元数据、堆栈跟踪和错误分类
- **认证**：基于JWT的认证，使用[gin-jwt](https://github.com/appleboy/gin-jwt)
- **调度器**：使用[robfig/cron](https://github.com/robfig/cron)的集成CRON定时任务调度器
- **迁移**：使用[goose](https://github.com/pressly/goose)进行数据库迁移
- **部署**：使用Docker和k3s进行容器化和编排

## 项目结构

```
├── cmd/                   # 应用程序入口点
│   ├── web-api/           # Web API服务器
│   ├── migrate/           # 数据库迁移工具
│   └── task/              # 用于定时、轮询和队列任务的任务运行器
├── config/                # 配置文件
│   ├── dev/               # 开发环境配置
│   └── prod/              # 生产环境配置
├── db/                    # 数据库相关文件
│   └── migrations/        # 数据库迁移文件
├── deploy/                # 部署配置
│   ├── config/            # 部署配置文件
│   ├── k3s/               # k3s部署清单
│   │   ├── cluster/       # 集群部署配置
│   │   └── single/        # 单节点部署配置
│   └── scripts/           # 部署自动化脚本
├── docs/                  # 文档
│   ├── deployment.md      # 详细部署指南（英文）
│   ├── deployment-zh.md   # 详细部署指南（中文）
├── internal/              # 私有应用代码
│   ├── api/               # API特定代码
│   │   └── http/          # HTTP API代码
│   │       └── web/       # Web API处理器和路由
│   │           ├── handler/   # API请求处理器
│   │           ├── middleware/# HTTP中间件组件
│   │           ├── types/     # 请求/响应结构
│   │           ├── service/   # Web API服务
│   │           └── router.go  # 路由定义
│   ├── enum/              # 枚举常量
│   ├── model/             # 领域模型
│   ├── repository/        # 数据访问层
│   ├── service/           # 服务层，协调处理器和用例之间的交互
│   ├── task/              # 任务管理
│   │   ├── poller/        # 轮询任务框架
│   │   ├── queue/         # 基于队列的任务框架
│   │   └── scheduler/     # 定时任务框架
│   └── usecase/           # 业务逻辑
├── pkg/                   # 公共库
│   ├── app/               # 应用程序框架
│   ├── binding/           # 请求绑定工具
│   ├── cache/             # 缓存
│   ├── config/            # 配置
│   ├── database/          # 数据库连接
│   ├── errs/              # 错误处理工具
│   ├── logger/            # 日志记录
│   ├── migrate/           # 数据库迁移工具
│   ├── mq/                # 消息队列
│   ├── server/            # HTTP服务器
│   └── utils/             # 工具函数
└── scripts/               # 自动化脚本
```

## 架构

### 层级职责

- **Handler层**：接收和解析HTTP请求，使用处理后的数据调用Service层。每个API端点使用专用的请求和响应结构。
- **Service层**：协调Handler和Usecase层之间的交互，处理数据转换，但不实现核心业务逻辑。
- **Usecase层**：包含独立于API层的核心业务逻辑。按照依赖倒置原则定义Repository接口。
- **Repository层**：管理数据访问和数据库交互。

### 中间件系统

应用程序包含多个中间件组件：

- **错误中间件**：API响应的集中错误处理，提供一致的错误格式
- **恢复中间件**：使用zap进行结构化日志记录的panic恢复
- **超时中间件**：请求超时强制控制
- **认证中间件**：基于JWT的认证
- **管理员专用中间件**：基于角色的管理员路由授权
- **CORS中间件**：跨域资源共享策略执行，采用生产就绪的安全默认设置

### 错误处理系统

应用程序实现了全面的错误处理系统：

- **错误类型**：业务错误、验证错误和内部错误
- **错误元数据**：支持额外的错误上下文
- **堆栈跟踪**：内部错误的自动堆栈跟踪捕获
- **错误分类**：错误分类及相应的HTTP状态码
- **结构化日志**：详细的错误日志记录，格式一致

### 任务系统架构

应用程序包含一个健壮的任务系统，具有三种任务执行模型：

- **定时任务**：基于CRON的任务，按指定间隔执行（例如日报、清理作业）
- **轮询任务**：以固定间隔运行的任务，持续检查条件或数据变化
- **队列任务**：通过RabbitMQ处理的异步任务，用于后台处理和工作负载分配

每种任务类型都遵循一致的注册和执行模式，使添加新任务变得容易，同时确保适当的生命周期管理和错误处理。

## 认证系统

应用程序实现了基于JWT的认证系统，具有以下特点：

- **短期令牌**：默认情况下，访问令牌在30分钟后过期，增强安全性
- **令牌刷新**：支持在可配置的时间窗口内刷新令牌（默认7天）
- **无状态设计**：无服务器端会话存储，完美适用于水平扩展
- **RESTful实现**：通过Authorization头传递令牌

前端应用程序可以通过以下端点与认证系统集成：

- `POST /api/web/v1/login` - 用于用户认证
- `GET /api/web/v1/refresh_token` - 用于刷新过期令牌

## API响应格式

所有API响应都遵循一致的结构：

```json
{
  "code": 200,           // HTTP状态码
  "message": "Success",  // 人类可读消息
  "data": {},            // 响应负载（成功时）
  "meta": {}             // 附加元数据（例如分页信息）
}
```

错误响应保持相同的结构：

```json
{
  "code": 400,                 // 错误码
  "message": "Validation error", // 错误消息
  "data": null,                // 错误时没有数据
  "meta": null                 // 错误时没有元数据
}
```

## 入门指南

### 前提条件

- Go 1.21或更高版本
- PostgreSQL
- Redis（可选）
- RabbitMQ（可选）

### 安装

1. 克隆仓库

   ```bash
   git clone https://github.com/xelarion/go-layout.git
   cd go-layout
   ```

2. 安装依赖

   ```bash
   go mod tidy
   ```

3. 设置环境变量（使用config/dev中的样例作为起点）

4. 运行数据库迁移

   ```bash
   # 最简单的方法，使用默认选项
   make db

   # 替代命令
   make migrate

   # 或直接使用CLI并提供更多选项
   go run cmd/migrate/main.go up
   go run cmd/migrate/main.go -dir=db/migrations -verbose status
   ```

5. 启动API服务器

   ```bash
   go run cmd/web-api/main.go
   ```

6. 启动任务运行器，带有所需组件（所有标志都是可选的）

   ```bash
   go run cmd/task/main.go --scheduler --poller --queue
   ```

### Docker部署

1. 构建Docker镜像

   ```bash
   docker build -t go-layout .
   ```

2. 使用Docker Compose运行

   ```bash
   docker-compose up -d
   ```

### 生产环境部署

有关详细的生产环境部署说明，请参阅我们的[部署指南](docs/deployment-zh.md)。

项目包含用于部署到单节点和多节点k3s集群的配置：

```bash
# 部署到单节点k3s环境
make deploy-single

# 部署到k3s集群环境
make deploy-cluster

# 部署到特定服务器（在deploy/config/servers.yaml中定义）
make deploy-server SERVER=servername
```

## API端点

API提供以下端点：

- **认证**
  - `POST /api/web/v1/login` - 登录并获取JWT令牌
  - `GET /api/web/v1/refresh_token` - 刷新JWT令牌
  - `GET /api/web/v1/captcha` - 获取登录验证码

- **用户管理**
  - `GET /api/web/v1/profile` - 获取当前用户个人资料（需要认证）
  - `PUT /api/web/v1/profile` - 更新当前用户个人资料（需要认证）
  - `POST /api/web/v1/users` - 创建新用户（需要管理员角色）
  - `GET /api/web/v1/users/:id` - 通过ID获取用户（需要管理员角色）
  - `PUT /api/web/v1/users/:id` - 更新用户（需要管理员角色）
  - `PATCH /api/web/v1/users/:id/enabled` - 更新用户启用状态（需要管理员角色）
  - `DELETE /api/web/v1/users/:id` - 删除用户（需要管理员角色）
  - `GET /api/web/v1/users` - 带分页和筛选的用户列表（需要管理员角色）

- **系统**
  - `GET /health` - 健康检查端点

## 许可证

[MIT](LICENSE)

## 贡献

1. Fork仓库
2. 创建特性分支
3. 提交您的更改
4. 推送到分支
5. 创建新的Pull Request

## 数据库迁移

本项目使用 [Goose](https://github.com/pressly/goose) 进行数据库迁移管理。迁移文件使用 SQL 编写，存储在 `db/migrations` 目录中。

### 迁移版本控制策略

我们采用 Goose 的混合版本控制方法来处理团队环境中的迁移：

1. **开发环境**：开发过程中的迁移文件自动使用时间戳命名（如 `20240628120000_add_users.sql`），这有助于避免多个开发人员同时创建迁移时的版本冲突。

2. **生产环境**：在部署到生产环境之前，迁移会被转换为顺序版本编号，同时保留原始顺序，确保在生产环境中可预测地应用。

这种混合方法既允许灵活开发，又保持了生产环境中的严格顺序。

### 本地运行迁移

您可以使用以下 Makefile 命令运行数据库迁移：

```bash
# 应用所有待处理的迁移
make migrate

# 检查迁移状态
make migrate-status

# 回滚最后一次迁移
make migrate-down

# 回滚所有迁移
make migrate-reset

# 创建新的迁移文件（自动使用时间戳）
make migrate-create NAME=add_users_table

# 修复迁移版本（将时间戳转换为顺序编号）
make migrate-fix

# 打印当前迁移版本
make migrate-version

# 回滚并重新应用最新的迁移
make migrate-redo

# 迁移到指定版本
make migrate-up-to VERSION=20240628120000

# 回滚到指定版本
make migrate-down-to VERSION=20240628120000
```

在本地开发期间，系统允许无序迁移，使团队成员可以独立工作而不会产生版本冲突。所有迁移通过正则表达式模式进行验证，以防止SQL注入。

### 生产环境迁移

在生产环境中，迁移由 Kubernetes Job 处理。在部署新版本时，应该在部署应用程序之前运行迁移：

```bash
# 构建迁移镜像（自动修复版本编号）
make build-migrate

# 在单节点 k3s 环境中运行迁移
make deploy-migrate

# 在 k3s 集群环境中运行迁移
make deploy-migrate-cluster
```

迁移作业作为Kubernetes作业运行，设置了`restartPolicy: Never`和`backoffLimit: 3`。它执行一次迁移操作后就完成。您可以使用以下命令检查作业状态和日志：

```bash
kubectl get jobs -n go-layout
kubectl logs job/db-migrate -n go-layout
```

### CI/CD 集成

对于 CI/CD 流水线，使用 `migrate-ci` 命令准备用于生产部署的迁移：

```bash
# 将时间戳版本转换为用于生产部署的顺序版本
make migrate-ci
```

这应该在构建迁移镜像之前在 CI 流水线中运行。`build-migrate`命令也会自动运行`migrate-fix`，确保在生产环境中使用顺序版本编号。

### 迁移文件

迁移文件遵循 Goose 命名约定：

```
{版本号}_{描述}.sql
```

每个迁移文件包含两个部分：
- `-- +goose Up`：升级时执行的命令
- `-- +goose Down`：回滚迁移时执行的命令

示例迁移文件：

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

### 使用Docker进行迁移

项目包含专用的`Dockerfile.migrate`用于将迁移工具容器化。该容器的入口点配置为运行迁移工具，默认显示迁移状态。在Kubernetes作业中，我们会覆盖这一设置以运行`up`命令。

```bash
# 在本地运行迁移容器（默认显示状态）
docker run go-layout-migrate

# 运行特定迁移命令
docker run go-layout-migrate up
docker run go-layout-migrate down
```
