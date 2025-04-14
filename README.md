# Go Layout

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
- **Database**: PostgreSQL (configurable to switch to MySQL)
- **Cache**: Redis (pluggable, can be removed if not needed)
- **Message Queue**: RabbitMQ (pluggable, can be removed if not needed)
- **Logging**: Structured logging with [zap](https://github.com/uber-go/zap)
- **Configuration**: Environment-based configuration using [github.com/caarlos0/env/v11](https://github.com/caarlos0/env)
- **Error Handling**: Custom error package with metadata, stack traces, and error categorization
- **Authentication**: JWT-based authentication using [gin-jwt](https://github.com/appleboy/gin-jwt)
- **Scheduler**: Integrated CRON scheduler for timed tasks using [robfig/cron](https://github.com/robfig/cron)
- **Migration**: Database migrations using [goose](https://github.com/pressly/goose)
- **Deployment**: Containerization and orchestration using Docker and k3s

## Project Structure

```
├── cmd/                           # Application entry points
│   ├── web-api/                   # Web API server
│   ├── migrate/                   # Database migration tool
│   └── task/                      # Task runner for scheduled, polling, and queue tasks
├── config/                        # Configuration files
│   ├── dev/                       # Development environment configs
│   └── prod/                      # Production environment configs
├── db/                            # Database related files
│   └── migrations/                # Database migration files
├── deploy/                        # Deployment configurations
│   ├── config/                    # Deployment config files
│   ├── k3s/                       # k3s deployment manifests
│   │   ├── cluster/               # Cluster deployment configurations
│   │   └── single/                # Single-node deployment configurations
│   └── scripts/                   # Deployment automation scripts
├── docs/                          # Documentation
│   ├── deployment.md              # Detailed deployment guide (English)
│   ├── deployment-zh.md           # Detailed deployment guide (Chinese)
├── internal/                      # Private application code
│   ├── api/                       # API-specific code
│   │   └── http/                  # HTTP API code
│   │       └── web/               # Web API handlers and routes
│   │           ├── handler/       # API request handlers
│   │           ├── middleware/    # HTTP middleware components
│   │           ├── types/         # Request/response structures
│   │           ├── service/       # Web API services
│   │           ├── swagger/       # Swagger documentation
│   │           └── router.go      # Route definitions
│   ├── enum/                      # Enumeration constants
│   ├── model/                     # Domain models
│   │   └── gen/                   # Generated models
│   ├── repository/                # Data access layer
│   ├── service/                   # Service layer, coordinates between handlers and usecases
│   ├── task/                      # Task management
│   │   ├── poller/                # Polling task framework
│   │   ├── queue/                 # Queue-based task framework
│   │   └── scheduler/             # Scheduled task framework
│   └── usecase/                   # Business logic
├── pkg/                           # Public libraries
│   ├── app/                       # Application framework
│   ├── binding/                   # Request binding tools
│   ├── cache/                     # Caching
│   ├── config/                    # Configuration
│   ├── database/                  # Database connections
│   ├── errs/                      # Error handling tools
│   ├── logger/                    # Logging
│   ├── migrate/                   # Database migration tools
│   ├── mq/                        # Message queue
│   ├── server/                    # HTTP server
│   └── utils/                     # Utility functions
├── tools/                         # Development tools
│   ├── gen/                       # Code generation tools
│   └── swagger_autocomment/       # Swagger comment generation tool
└── scripts/                       # Automation scripts
```

## Architecture

### Layer Responsibilities

- **Handler Layer**: Receives and parses HTTP requests, calls Service layer with processed data. Each API endpoint uses dedicated request and response structures.
- **Service Layer**: Coordinates between Handler and Usecase layers, handles data transformations but doesn't implement core business logic.
- **Usecase Layer**: Contains core business logic independent of the API layer. Defines Repository interfaces as per the dependency inversion principle.
- **Repository Layer**: Manages data access and database interactions.

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
  "code": 200,
  "message": "Success",
  "data": {},
  "meta": {}
}
```

Error responses maintain the same structure:

```json
{
  "code": 400,
  "message": "Validation error"
}
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL
- Redis (optional)
- RabbitMQ (optional)

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

8. Start the task runner with desired components (all flags are optional)

   ```bash
   go run ./cmd/task --scheduler --poller --queue
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

## Model Generation

This project uses GORM's model generation tool to automatically create Go structs from database tables. The generated models support various PostgreSQL data types, including arrays and JSON.

### Supported Data Types

The model generator has enhanced support for PostgreSQL data types:

- **Basic Types**: All standard PostgreSQL types (integer, text, etc.)
- **Array Types**: Arrays are mapped to appropriate Go types:
  - `character varying[]`, `varchar[]`, `text[]` → `pq.StringArray`
  - `integer[]`, `int[]` → `pq.Int32Array`
  - `bigint[]` → `pq.Int64Array`
  - `boolean[]` → `pq.BoolArray`
  - `numeric[]`, `decimal[]` → `pq.Float64Array`
- **JSON Types**: `json` and `jsonb` are mapped to `datatypes.JSON` from `github.com/go-gorm/datatypes`
- **Special Types**:
  - `uuid` → `datatypes.UUID`
  - `date` → `datatypes.Date`
  - `time` → `datatypes.Time`

### Running the Model Generator

You can generate models using the following commands:

```bash
# Generate models for all tables in the database
make gen-models

# Generate model for a specific table (useful for team development)
make gen-model TABLE=users
```

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

## License

[MIT](LICENSE)

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
