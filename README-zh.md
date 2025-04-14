# Go 项目布局

一个可扩展、高性能、高可用性的 Go 语言 Web 应用程序模板。

## 功能特性

- **分层架构**：Handler → Service → Usecase → Repository 模式，具有清晰的关注点分离
- **高性能**：优化的数据库查询、高效的缓存策略和性能监控
- **高可用性**：无状态设计，便于水平扩展和集群部署
- **可扩展性**：模块化设计，支持水平和垂直扩展
- **代码整洁**：结构良好、可维护且可扩展的代码库
- **安全性**：JWT 认证、输入验证和常见漏洞防护
- **错误处理**：全面的错误处理系统，包含结构化日志和自定义错误类型
- **中间件**：模块化中间件组件，用于请求处理、错误处理、请求超时控制和 CORS
- **任务管理**：集成支持计划任务、轮询任务和基于队列的异步处理

## 技术栈

- **编程语言**：Go
- **Web 框架**：[Gin](https://github.com/gin-gonic/gin)
- **ORM**：[GORM](https://gorm.io/)
- **数据库**：PostgreSQL（可配置切换为 MySQL）
- **缓存**：Redis（可插拔，不需要时可移除）
- **消息队列**：RabbitMQ（可插拔，不需要时可移除）
- **日志记录**：使用 [zap](https://github.com/uber-go/zap) 进行结构化日志记录
- **配置管理**：使用 [github.com/caarlos0/env/v11](https://github.com/caarlos0/env) 进行环境变量配置
- **错误处理**：自定义错误包，支持元数据、堆栈跟踪和错误分类
- **认证系统**：使用 [gin-jwt](https://github.com/appleboy/gin-jwt) 实现基于 JWT 的认证
- **调度器**：使用 [robfig/cron](https://github.com/robfig/cron) 集成 CRON 调度器，用于定时任务
- **数据库迁移**：使用 [goose](https://github.com/pressly/goose) 进行数据库迁移
- **部署方案**：使用 Docker 和 k3s 进行容器化和编排

## 项目结构

```
├── cmd/                           # 应用程序入口点
│   ├── web-api/                   # Web API 服务器
│   ├── migrate/                   # 数据库迁移工具
│   └── task/                      # 任务运行器（定时任务、轮询任务和队列任务）
├── config/                        # 配置文件
│   ├── dev/                       # 开发环境配置
│   └── prod/                      # 生产环境配置
├── db/                            # 数据库相关文件
│   └── migrations/                # 数据库迁移文件
├── deploy/                        # 部署配置
│   ├── config/                    # 部署配置文件
│   ├── k3s/                       # k3s 部署清单
│   │   ├── cluster/               # 集群部署配置
│   │   └── single/                # 单节点部署配置
│   └── scripts/                   # 部署自动化脚本
├── docs/                          # 文档
│   ├── deployment.md              # 详细部署指南（英文）
│   ├── deployment-zh.md           # 详细部署指南（中文）
├── internal/                      # 私有应用代码
│   ├── api/                       # API 专用代码
│   │   └── http/                  # HTTP API 代码
│   │       └── web/               # Web API 处理器和路由
│   │           ├── handler/       # API 请求处理器
│   │           ├── middleware/    # HTTP 中间件组件
│   │           ├── types/         # 请求/响应结构
│   │           ├── service/       # Web API 服务
│   │           ├── swagger/       # Swagger 文档
│   │           └── router.go      # 路由定义
│   ├── enum/                      # 枚举常量
│   ├── model/                     # 领域模型
│   │   └── gen/                   # 生成的模型
│   ├── repository/                # 数据访问层
│   ├── service/                   # 服务层，协调处理器和用例层之间的交互
│   ├── task/                      # 任务管理
│   │   ├── poller/                # 轮询任务框架
│   │   ├── queue/                 # 基于队列的任务框架
│   │   └── scheduler/             # 计划任务框架
│   └── usecase/                   # 业务逻辑
├── pkg/                           # 公共库
│   ├── app/                       # 应用程序框架
│   ├── binding/                   # 请求绑定工具
│   ├── cache/                     # 缓存
│   ├── config/                    # 配置
│   ├── database/                  # 数据库连接
│   ├── errs/                      # 错误处理工具
│   ├── logger/                    # 日志记录
│   ├── migrate/                   # 数据库迁移工具
│   ├── mq/                        # 消息队列
│   ├── server/                    # HTTP 服务器
│   └── utils/                     # 实用工具函数
├── tools/                         # 开发工具
│   ├── gen/                       # 代码生成工具
│   └── swagger_autocomment/       # Swagger 注释生成工具
└── scripts/                       # 自动化脚本
```

## 架构设计

### 层级职责

- **处理器层（Handler）**：接收并解析 HTTP 请求，使用处理后的数据调用服务层。每个 API 端点使用专用的请求和响应结构。
- **服务层（Service）**：协调处理器层和用例层之间的交互，处理数据转换，但不实现核心业务逻辑。
- **用例层（Usecase）**：包含独立于 API 层的核心业务逻辑。根据依赖倒置原则定义仓库接口。
- **仓库层（Repository）**：管理数据访问和数据库交互。

### 中间件系统

应用程序包含多个中间件组件：

- **错误中间件**：API 响应的集中式错误处理，提供一致的错误格式
- **恢复中间件**：使用 zap 进行结构化日志记录的 panic 恢复
- **超时中间件**：请求超时强制执行
- **认证中间件**：基于 JWT 的认证
- **CORS 中间件**：跨源资源共享策略实施，具有生产环境安全默认设置

### 错误处理系统

应用程序实现了全面的错误处理系统：

- **错误类型**：业务错误、验证错误和内部错误
- **错误元数据**：支持附加错误上下文
- **堆栈追踪**：自动为内部错误捕获堆栈跟踪
- **错误分类**：对错误进行分类，对应相应的 HTTP 状态码
- **结构化日志**：详细的错误日志记录，格式一致

### 任务系统架构

应用程序包含一个健壮的任务系统，具有三种任务执行模型：

- **计划任务**：基于 CRON 的任务，在指定时间间隔执行（例如，每日报告、清理作业）
- **轮询任务**：以固定间隔运行的任务，持续检查条件或数据变化
- **队列任务**：通过 RabbitMQ 处理的异步任务，用于后台处理和工作负载分配

每种任务类型都遵循一致的注册和执行模式，使添加新任务变得容易，同时确保适当的生命周期管理和错误处理。

## 认证系统

应用程序实现了基于 JWT 的认证系统，具有以下特点：

- **短寿命令牌**：默认情况下，访问令牌在 30 分钟后过期，增强安全性
- **令牌刷新**：支持在可配置时间窗口内刷新令牌（默认为 7 天）
- **无状态设计**：无服务器端会话存储，非常适合水平扩展
- **RESTful 实现**：通过 Authorization 标头传递令牌

前端应用程序可以通过以下端点与认证系统集成：

- `POST /api/web/v1/login` - 用户认证
- `GET /api/web/v1/refresh_token` - 刷新过期令牌

## API 响应格式

所有 API 响应都遵循一致的结构：

```json
{
  "code": 200,
  "message": "Success",
  "data": {},
  "meta": {}
}
```

错误响应保持相同的结构：

```json
{
  "code": 400,
  "message": "Validation error"
}
```

## 入门指南

### 先决条件

- Go 1.21 或更高版本
- PostgreSQL
- Redis（可选）
- RabbitMQ（可选）

### 安装

1. 克隆存储库

   ```bash
   git clone https://github.com/xelarion/go-layout.git
   cd go-layout
   ```

2. 安装依赖项

   ```bash
   go mod tidy
   ```

3. 设置环境变量（使用 config/dev 中的示例作为起点）

4. 运行数据库迁移

   ```bash
   # 应用所有待处理的迁移
   make migrate

   # 检查迁移状态
   make migrate-status

   # 或者直接使用 CLI 获取更多选项
   go run cmd/migrate/main.go up
   go run cmd/migrate/main.go -dir=db/migrations -verbose status
   ```

5. 生成数据库模型

   ```bash
   # 生成所有表的模型
   make gen-models

   # 生成特定表的模型（适用于团队开发）
   make gen-model TABLE=users
   ```

6. 生成 API 文档（可选）

   ```bash
   # 首先为处理器生成 Swagger 注释
   make swagger-comment

   # 然后生成 Swagger 文档
   make swagger-docs

   # 或者使用单个命令完成两者
   make swagger-all
   ```

7. 启动 API 服务器

   ```bash
   go run ./cmd/web-api
   ```

8. 启动任务运行器，选择所需组件（所有标志都是可选的）

   ```bash
   go run ./cmd/task --scheduler --poller --queue
   ```

### Docker 部署

1. 构建 Docker 镜像

   ```bash
   # 构建所有服务
   make build

   # 或构建单个服务
   make build-web-api
   make build-task
   make build-migrate
   ```

2. 使用 Docker Compose 运行

   ```bash
   docker-compose up -d
   ```

### 生产环境部署

有关详细的生产部署说明，请参阅我们的[部署指南](docs/deployment-zh.md)。

项目包含用于部署到单节点和多节点 k3s 集群的配置：

```bash
# 首先运行迁移（在部署服务之前运行迁移很重要）
make deploy-migrate

# 部署到单节点 k3s 环境
make deploy-single

# 部署到 k3s 集群环境
make deploy-cluster

# 使用 k3s 部署脚本进行部署
make deploy-k3s
```

## 数据库迁移

本项目使用 [Goose](https://github.com/pressly/goose) 进行数据库迁移管理。迁移文件以 SQL 编写，存储在 `db/migrations` 目录中。

### 迁移版本控制策略

我们使用 Goose 的混合版本控制方法来处理团队环境中的迁移：

1. **开发环境**：在开发过程中，迁移文件自动以时间戳命名（例如，`20240628120000_add_users.sql`），这有助于避免多个开发人员同时创建迁移时的版本冲突。

2. **生产环境**：在部署到生产环境之前，迁移会被转换为顺序版本号，同时保留原始顺序，确保在生产环境中可预测地应用。

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
make migrate-create NAME=create_users

# 修复迁移版本（将时间戳转换为顺序编号）
make migrate-fix

# 打印当前迁移版本
make migrate-version

# 回滚并重新应用最新迁移
make migrate-redo

# 迁移到特定版本
make migrate-up-to VERSION=20240628120000

# 回滚到特定版本
make migrate-down-to VERSION=20240628120000
```

### 生产环境迁移

在生产环境中，迁移由 Kubernetes Job 处理。部署新版本时，应在部署应用程序之前运行迁移：

```bash
# 构建迁移镜像（自动修复版本号）
make build-migrate

# 在单节点 k3s 环境中运行迁移
make deploy-migrate

# 在 k3s 集群环境中运行迁移
make deploy-migrate-cluster
```

迁移作业作为 Kubernetes 作业运行，具有 `restartPolicy: Never` 和 `backoffLimit: 3`。

## 模型生成

本项目使用 GORM 的模型生成工具从数据库表自动创建 Go 结构体。生成的模型支持各种 PostgreSQL 数据类型，包括数组和 JSON。

### 支持的数据类型

模型生成器对 PostgreSQL 数据类型有增强支持：

- **基本类型**：所有标准 PostgreSQL 类型（整数、文本等）
- **数组类型**：数组被映射到适当的 Go 类型：
  - `character varying[]`、`varchar[]`、`text[]` → `pq.StringArray`
  - `integer[]`、`int[]` → `pq.Int32Array`
  - `bigint[]` → `pq.Int64Array`
  - `boolean[]` → `pq.BoolArray`
  - `numeric[]`、`decimal[]` → `pq.Float64Array`
- **JSON 类型**：`json` 和 `jsonb` 被映射到 `github.com/go-gorm/datatypes` 的 `datatypes.JSON`
- **特殊类型**：
  - `uuid` → `datatypes.UUID`
  - `date` → `datatypes.Date`
  - `time` → `datatypes.Time`

### 运行模型生成器

您可以使用以下命令生成模型：

```bash
# 生成数据库中所有表的模型
make gen-models

# 生成特定表的模型（适用于团队开发）
make gen-model TABLE=users
```

### 使用生成的模型

生成的模型位于 `internal/model/gen` 目录中。对于每个生成的模型，您应该在 `internal/model` 目录中创建一个相应的扩展模型。

扩展模型示例：

```go
package model

import (
    "github.com/xelarion/go-layout/internal/model/gen"
)

// User 表示用户模型。
type User struct {
    gen.User
}
```

这种方法允许您向模型添加自定义方法和属性，同时保持从数据库模式重新生成基本模型的能力。

## 许可证

[MIT](LICENSE)

## 贡献

1. Fork 存储库
2. 创建您的功能分支
3. 提交您的更改
4. 推送到分支
5. 创建新的 Pull Request
