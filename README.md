# Go Layout

[English](README.md) | [中文](README-zh.md)

A scalable, high-performance, high-availability web application template built with Go.

## Features

- **Layered Architecture**: Handler → Service → Usecase → Repository pattern with clear separation of concerns
- **High Performance**: Optimized database queries, efficient caching strategy, and performance monitoring
- **High Availability**: Stateless design for easy horizontal scaling and cluster deployment
- **Scalability**: Modular design supporting both horizontal and vertical scaling
- **Clean Code**: Well-structured, maintainable, and extensible codebase
- **Security**: JWT authentication, input validation, and protection against common vulnerabilities
- **Error Handling**: Comprehensive error handling system with structured logging and custom error types
- **Middleware**: Modular middleware components for request processing, error handling, request timeout control, and CORS
- **Task Management**: Integrated support for scheduled tasks, polling tasks, and queue-based asynchronous processing

## Technology Stack

- **Programming Language**: Go
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Logging**: Structured logging with [zap](https://github.com/uber-go/zap)
- **Configuration**: Environment-based configuration using [github.com/caarlos0/env/v11](https://github.com/caarlos0/env)
- **Error Handling**: Custom error package with metadata, stack traces, and error categorization
- **Authentication**: JWT-based authentication using [gin-jwt](https://github.com/appleboy/gin-jwt)
- **Dependency Injection**: Clean dependency injection using [Google Wire](https://github.com/google/wire)
- **Scheduler**: Integrated CRON scheduler for timed tasks using [robfig/cron](https://github.com/robfig/cron)
- **Migration**: Database migrations using [goose](https://github.com/pressly/goose)
- **Deployment**: Containerization and orchestration using Docker and k3s

## Project Structure

```markdown
├── cmd/                           # Application entry points
│   ├── web-api/                   # Web API server
│   ├── migrate/                   # Database migration tool
│   └── task/                      # Task runner
├── config/                        # Configuration files
│   ├── dev/                       # Development environment configs
│   └── prod/                      # Production environment configs
├── migrations/                    # Database migration files
├── deploy/                        # Deployment configurations
│   ├── config/                    # Deployment config files
│   ├── k3s/                       # k3s deployment manifests
│   │   ├── cluster/               # Cluster deployment configurations
│   │   └── single/                # Single-node deployment configurations
│   └── scripts/                   # Deployment automation scripts
├── docs/                          # Documentation
├── internal/                      # Private application code
│   ├── api/                       # API-specific code
│   │   └── http/                  # HTTP API code
│   │       └── web/               # Web API handlers and routes
│   │           ├── handler/       # API request handlers
│   │           ├── middleware/    # HTTP middleware components
│   │           ├── service/       # Web API service layer
│   │           ├── swagger/       # Swagger documentation
│   │           ├── types/         # Request/response structures
│   │           └── router.go      # Route definitions
│   ├── enum/                      # Enumeration constants
│   ├── infra/                     # Infrastructure layer
│   │   ├── cache/                 # Cache implementations (Redis)
│   │   ├── config/                # Configuration loading
│   │   ├── database/              # Database connections
│   │   ├── logger/                # Logger initialization
│   │   ├── migrate/               # Migration infrastructure
│   │   ├── mq/                    # Message queue (RabbitMQ)
│   │   └── server/                # Server implementations
│   │       └── http/              # HTTP server
│   ├── model/                     # Domain models
│   │   └── gen/                   # Generated models
│   ├── permission/                # Authorization system
│   ├── repository/                # Data access layer
│   ├── task/                      # Task management
│   │   ├── dependencies.go        # Task dependencies setup
│   │   ├── poller/                # Polling task framework
│   │   │   └── tasks/             # Polling task implementations
│   │   ├── queue/                 # Queue-based task framework
│   │   │   └── tasks/             # Queue task implementations
│   │   └── scheduler/             # Scheduled task framework
│   │       └── tasks/             # Scheduled task implementations
│   ├── usecase/                   # Business logic layer
│   └── util/                      # Utility functions
├── pkg/                           # Public libraries
│   ├── app/                       # Application framework
│   │   ├── app.go                 # Core application lifecycle
│   │   └── options.go             # Application options
│   ├── binding/                   # Request binding utilities
│   └── errs/                      # Error handling package
└── tools/                         # Development tools
```

## Architecture

### Layer Responsibilities

- **Handler Layer**: Receives and parses HTTP requests, calls Service layer with processed data. Each API endpoint uses dedicated request and response structures.
- **Service Layer**: Coordinates between Handler and Usecase layers, handles data transformations but doesn't implement core business logic.
- **Usecase Layer**: Contains core business logic independent of the API layer. Defines Repository interfaces as per the dependency inversion principle.
- **Repository Layer**: Manages data access and database interactions.

### Dependency Injection

The application uses Google Wire for dependency injection with a modular provider set approach:

- **Provider Sets**: Each layer (repository, usecase, service, handler) maintains its own provider set, making dependencies explicit and easier to manage
- **Auto-Generated Code**: Wire automatically generates the dependency injection code, eliminating manual wiring
- **Modularity**: Adding new components only requires updating the relevant provider set and regenerating the wire_gen.go file

### Permission System

The application implements a role-based permission system for access control:

- **Permission Definition**: Permissions are defined as constants in the `permission` package (e.g., `user:list`, `role:update`)
- **Permission Tree**: Permissions are organized in a hierarchical structure exposed via the `/api/v1/permissions/tree` endpoint
- **Permission Usage**: API endpoints are protected using middleware:

  ```go
  // Single permission check
  router.GET("/users", permMW.Check(permission.UserList), handler.ListUsers)

  // Multiple permissions check (any of them grants access)
  router.GET("/users/:id", permMW.Check(permission.UserDetail, permission.UserUpdate), handler.GetUser)
  ```

### Middleware System

The application includes multiple middleware components:

- **Error Middleware**: Centralized error handling for API responses, providing consistent error formatting
- **Recovery Middleware**: Panic recovery with structured logging using zap
- **Timeout Middleware**: Request timeout enforcement
- **Authentication Middleware**: JWT-based authentication
- **CORS Middleware**: Cross-Origin Resource Sharing policy enforcement with production-ready secure defaults

### Error Handling System

The application implements a comprehensive error handling system:

- **Error Types**: Business errors, validation errors, and internal errors
- **Error Metadata**: Support for additional error context
- **Stack Traces**: Automatic stack trace capture for internal errors
- **Error Categorization**: Categorization of errors with corresponding HTTP status codes
- **Structured Logging**: Detailed error logging with consistent formatting

### Transaction Management

The application implements a clean transaction management approach:

- **Transaction Interface**: A simple interface in the usecase layer that defines transaction boundaries
- **Context Propagation**: Transactions are passed through context for consistent state
- **Dependency Injection**: Transaction manager is injected into usecases for testability

Example transaction usage:

```go
// Create creates a new department.
func (uc *DepartmentUseCase) Create(ctx context.Context, params CreateDepartmentParams) (uint, error) {
	department := &model.Department{
		Department: gen.Department{
			Name:        params.Name,
			Description: params.Description,
			Enabled:     params.Enabled,
		},
	}
	err := uc.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := uc.departmentRepo.Create(ctx, department); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return department.ID, nil
}
```

### Task System Architecture

The application contains a robust task system with three task execution models:

- **Scheduled Tasks**: CRON-based tasks that run at specified intervals (e.g., daily reports, cleanup jobs)
- **Polling Tasks**: Tasks that run at fixed intervals, continuously checking for conditions or data changes
- **Queue Tasks**: Asynchronous tasks processed through RabbitMQ for background processing and workload distribution

Each task type follows a consistent registration and execution pattern, making it easy to add new tasks while ensuring proper lifecycle management and error handling.

## Authentication System

The application implements a JWT-based authentication system with the following features:

- **Short-lived Tokens**: Access tokens expire after 30 minutes by default, enhancing security
- **Token Refreshing**: Support for refreshing tokens within a configurable time window (7 days by default)
- **Stateless Design**: No server-side session storage, perfect for horizontal scaling
- **RESTful Implementation**: Tokens passed via Authorization header

Frontend applications can integrate with the authentication system through the following endpoints:

- `POST /api/web/v1/login` - User authentication
- `GET /api/web/v1/refresh_token` - Refresh expired tokens

## API Response Format

All API responses follow a consistent structure:

```json
{
  "code": 0,
  "message": "Success",
  "data": {},
  "meta": {}
}
```

Error responses maintain the same structure:

```json
{
  "code": 1000,
  "message": "Bad request"
}
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL
- Redis
- RabbitMQ

### Installation

1. Clone the repository

   ```bash
   git clone https://github.com/xelarion/go-layout.git
   cd go-layout
   ```

2. Install dependencies

   ```bash
   go mod tidy
   ```

3. Set up environment variables (use the samples in config/dev as a starting point)

4. Run database migrations

   ```bash
   # Apply all pending migrations
   make migrate

   # Check migration status
   make migrate-status

   # Or use the CLI directly for more options
   go run cmd/migrate/main.go up
   go run cmd/migrate/main.go -dir=migrations -verbose status
   ```

5. Generate database models

   ```bash
   # Generate models for all tables
   make gen-models

   # Generate model for a specific table (useful for team development)
   make gen-model TABLE=users
   ```

6. Generate API documentation (optional)

   ```bash
   # First generate Swagger comments for handlers
   make swagger-comment

   # Then generate Swagger documentation
   make swagger-docs

   # Or use a single command to do both
   make swagger-all
   ```

7. Start the API server

   ```bash
   go run ./cmd/web-api
   ```

8. Start the task runner with desired components

   ```bash
   go run ./cmd/task
   ```

### Docker Deployment

1. Build Docker images

   ```bash
   # Build all services
   make build

   # Or build individual services
   make build-web-api
   make build-task
   make build-migrate
   ```

2. Run with Docker Compose

   ```bash
   docker-compose up -d
   ```

### Production Deployment

For detailed production deployment instructions, see our [deployment guide](docs/deployment.md).

The project includes configurations for deployment to both single-node and multi-node k3s clusters:

```bash
# First run migrations (it's important to run migrations before deploying services)
make deploy-migrate

# Deploy to single-node k3s environment
make deploy-single

# Deploy to k3s cluster environment
make deploy-cluster

# Deploy using k3s deployment script
make deploy-k3s
```

## Code Generation

This project uses various code generation tools to improve development efficiency.

### Running Code Generation Tools

```bash
# Generate wire dependency injection code
make gen-wire

# Generate wire code for web-api service only
make gen-wire-web

# Generate models from database schema
make gen-models

# Generate model for a specific table
make gen-model TABLE=users

# Generate swagger documentation
make swagger-docs

# Generate intelligent swagger comments
make swagger-comment ARGS="-silent"
```

### Working with Wire

Wire generates dependency injection code based on provider functions. To modify the dependency graph:

1. Update the relevant provider set in your layer (repository, usecase, service, etc.)
2. Run `make gen-wire` to regenerate the wire_gen.go file
3. The application will now use the updated dependency graph

### Working with Generated Models

Generated models are placed in the `internal/model/gen` directory. For each generated model, you should create a corresponding extended model in the `internal/model` directory.

Example of an extended model:

```go
package model

import (
    "github.com/xelarion/go-layout/internal/model/gen"
)

// User represents a user model.
type User struct {
    gen.User
}
```

This approach allows you to add custom methods and properties to the models while maintaining the ability to regenerate the base models from the database schema.

## Database Migrations

This project uses [Goose](https://github.com/pressly/goose) for database migration management. Migration files are written in SQL and stored in the `migrations` directory.

### Migration Versioning Strategy

We use Goose's hybrid versioning approach to handle migrations in a team environment:

1. **Development Environment**: During development, migration files are automatically named with timestamps (e.g., `20240628120000_add_users.sql`), which helps avoid version conflicts when multiple developers create migrations simultaneously.

2. **Production Environment**: Before deployment to production, migrations are converted to sequential version numbers while preserving the original order, ensuring predictable application in production environments.

### Running Migrations Locally

You can run database migrations using the following Makefile commands:

```bash
# Apply all pending migrations
make migrate

# Check migration status
make migrate-status

# Rollback the last migration
make migrate-down

# Rollback all migrations
make migrate-reset

# Create a new migration file (automatically using timestamp)
make migrate-create NAME=create_users

# Fix migration versions (convert timestamps to sequential numbers)
make migrate-fix

# Print current migration version
make migrate-version

# Rollback and reapply the latest migration
make migrate-redo

# Migrate to a specific version
make migrate-up-to VERSION=20240628120000

# Rollback to a specific version
make migrate-down-to VERSION=20240628120000
```

### Production Migrations

In production environments, migrations are handled by a Kubernetes Job. When deploying a new version, migrations should be run before deploying the application:

```bash
# Build migration image (automatically fixes version numbers)
make build-migrate

# Run migrations in single-node k3s environment
make deploy-migrate

# Run migrations in k3s cluster environment
make deploy-migrate-cluster
```

The migration job runs as a Kubernetes job with `restartPolicy: Never` and `backoffLimit: 3`.

## License

[MIT](LICENSE)

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
