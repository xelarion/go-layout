# Go Layout

一个使用Go构建的可扩展、高性能、高可用性的Web应用模板。

## 特性

- **分层架构**：采用Handler → Service → Usecase → Repository分层设计模式，实现关注点分离
- **高性能**：优化的数据库查询、高效的缓存策略和性能监控
- **高可用性**：无状态设计，便于水平扩展和集群部署
- **可扩展性**：模块化设计，支持水平和垂直扩展
- **代码整洁**：组织良好、可维护且可扩展的代码库
- **安全性**：JWT认证、输入验证和防御常见漏洞
- **任务管理**：集成支持定时任务、轮询任务和基于队列的异步处理

## 技术栈

- **编程语言**：Go
- **Web框架**：Gin
- **ORM**：Gorm
- **数据库**：PostgreSQL（可配置切换到MySQL）
- **缓存**：Redis（可插拔，如不需要可移除）
- **消息队列**：RabbitMQ（可插拔，如不需要可移除）
- **调度器**：基于robfig/cron的集成CRON任务调度器
- **部署**：Docker和k3s用于容器化和编排

## 项目结构

```go
.
├── cmd/                   # 应用入口点
│   ├── api/               # API服务器
│   └── task/              # 定时任务、轮询任务和队列任务的运行器
├── config/                # 配置文件
│   ├── dev/               # 开发环境配置
│   └── prod/              # 生产环境配置
├── db/                    # 数据库相关文件
│   └── migrations/        # 数据库迁移文件
├── internal/              # 私有应用代码
│   ├── api/               # API特定代码
│   │   ├── public/        # 公共API处理器和路由
│   │   └── web/           # Web API处理器和路由
│   ├── middleware/        # HTTP中间件
│   ├── model/             # 领域模型
│   ├── repository/        # 数据访问层
│   ├── service/           # 服务层，协调处理器和用例
│   ├── task/              # 任务管理
│   │   ├── poller/        # 轮询任务框架
│   │   ├── queue/         # 队列任务框架
│   │   └── scheduler/     # 定时任务框架
│   └── usecase/           # 业务逻辑
├── pkg/                   # 公共库
│   ├── app/               # 应用框架
│   ├── auth/              # 认证
│   ├── cache/             # 缓存
│   ├── config/            # 配置
│   ├── database/          # 数据库连接
│   ├── logger/            # 日志
│   ├── mq/                # 消息队列
│   └── server/            # HTTP服务器
└── scripts/               # 自动化脚本
```

## 架构

### 层级职责

- **Handler层**：接收和解析HTTP请求，处理后调用Service层
- **Service层**：协调Handler和Usecase层，处理数据转换
- **Usecase层**：包含核心业务逻辑，独立于API层
- **Repository层**：管理数据访问和数据库交互

### 任务系统架构

应用包含一个强大的任务系统，有三种任务执行模型：

- **定时任务**：基于CRON的任务，按指定间隔执行（如每日报告、清理作业）
- **轮询任务**：以固定间隔运行的任务，持续检查条件或数据变化
- **队列任务**：通过RabbitMQ处理的异步任务，用于后台处理和工作负载分配

每种任务类型都遵循一致的注册和执行模式，便于添加新任务，同时确保适当的生命周期管理和错误处理。

## 快速开始

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
go run cmd/migrate/main.go
```

5. 启动API服务器

```bash
go run cmd/api/main.go
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

## API端点

API提供以下端点：
- **认证**
  - `POST /api/v1/register` - 注册新用户
  - `POST /api/v1/login` - 登录并获取JWT令牌
- **用户管理**
  - `GET /api/v1/profile` - 获取当前用户资料（需要认证）
  - `GET /api/v1/users/:id` - 通过ID获取用户（需要认证）
  - `PUT /api/v1/users/:id` - 更新用户（需要认证，只能更新自己的资料或需要管理员权限）
  - `GET /api/v1/users` - 列出用户（需要管理员角色）
  - `DELETE /api/v1/users/:id` - 删除用户（需要管理员角色）

## 许可证

[MIT](LICENSE)

## 贡献
1. Fork仓库
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建新的Pull Request