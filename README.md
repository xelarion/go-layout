# Go Layout

A scalable, high-performance, and high-availability web application template built with Go.

## Features

- **Layered Architecture**: Clear separation of concerns with Handler → Service → Usecase → Repository layered design pattern
- **High Performance**: Optimized database queries, efficient caching strategies, and performance monitoring
- **High Availability**: Stateless design for easy horizontal scaling and cluster deployment
- **Scalability**: Modular design that supports both horizontal and vertical scaling
- **Clean Code**: Well-organized, maintainable, and extensible codebase
- **Security**: JWT authentication, input validation, and protection against common vulnerabilities
- **Error Handling**: Comprehensive error handling system with structured logging and custom error types
- **Middleware**: Modular middleware components for request processing, error handling, request timeouts, and CORS
- **Task Management**: Integrated support for scheduled tasks, polling tasks, and queue-based asynchronous processing

## Technology Stack

- **Programming Language**: Go
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: PostgreSQL (configurable to switch to MySQL)
- **Cache**: Redis (pluggable, can be removed if not needed)
- **Message Queue**: RabbitMQ (pluggable, can be removed if not needed)
- **Logging**: Structured logging with [zap](https://github.com/uber-go/zap)
- **Configuration**: Environment-based configuration with [github.com/caarlos0/env/v11](https://github.com/caarlos0/env)
- **Error Handling**: Custom error package with metadata, stack traces, and error categorization
- **Authentication**: JWT-based authentication with [gin-jwt](https://github.com/appleboy/gin-jwt)
- **Scheduler**: Integrated CRON-based task scheduler using [robfig/cron](https://github.com/robfig/cron)
- **Migration**: Database migration with [goose](https://github.com/pressly/goose)
- **Deployment**: Docker and k3s for containerization and orchestration

## Project Structure

```
├── cmd/                   # Application entry points
│   ├── web-api/           # Web API server
│   ├── migrate/           # Database migration tool
│   └── task/              # Task runner for scheduled, polling and queue tasks
├── config/                # Configuration files
│   ├── dev/               # Development environment configs
│   └── prod/              # Production environment configs
├── db/                    # Database related files
│   └── migrations/        # Database migration files
├── deploy/                # Deployment configuration
│   ├── config/            # Deployment configuration files
│   ├── k3s/               # k3s deployment manifests
│   │   ├── cluster/       # Cluster deployment configurations
│   │   └── single/        # Single node deployment configurations
│   └── scripts/           # Deployment automation scripts
├── docs/                  # Documentation
│   ├── deployment.md      # Detailed deployment guide
│   ├── deployment-zh.md   # Deployment guide in Chinese
├── internal/              # Private application code
│   ├── api/               # API-specific code
│   │   └── http/          # HTTP API code
│   │       └── web/       # Web API handlers and routers
│   │           ├── handler/   # API request handlers
│   │           ├── middleware/# HTTP middleware components
│   │           ├── types/     # Request/response structures
│   │           ├── service/   # Web API services
│   │           └── router.go  # Route definitions
│   ├── enum/              # Enumeration constants
│   ├── model/             # Domain models
│   ├── repository/        # Data access layer
│   ├── service/           # Services coordinating between handlers and usecases
│   ├── task/              # Task management
│   │   ├── poller/        # Polling tasks framework
│   │   ├── queue/         # Queue-based tasks framework
│   │   └── scheduler/     # Scheduled tasks framework
│   └── usecase/           # Business logic
├── pkg/                   # Public libraries
│   ├── app/               # Application framework
│   ├── binding/           # Request binding utilities
│   ├── cache/             # Caching
│   ├── config/            # Configuration
│   ├── database/          # Database connections
│   ├── errs/              # Error handling utilities
│   ├── logger/            # Logging
│   ├── migrate/           # Database migration utilities
│   ├── mq/                # Message queue
│   ├── server/            # HTTP server
│   └── utils/             # Utility functions
└── scripts/               # Scripts for automation
```

## Architecture

### Layer Responsibilities

- **Handler Layer**: Receives and parses HTTP requests, invokes the Service layer with processed data. Each API endpoint uses dedicated request and response structures.
- **Service Layer**: Coordinates between Handler and Usecase layers, handles data transformation without implementing core business logic.
- **Usecase Layer**: Contains core business logic, independent of the API layer. Defines Repository interfaces following dependency inversion principle.
- **Repository Layer**: Manages data access and database interactions.

### Middleware System

The application includes several middleware components:

- **Error Middleware**: Centralized error handling for API responses with consistent error formats
- **Recovery Middleware**: Panic recovery with structured logging using zap
- **Timeout Middleware**: Request timeout enforcement
- **Authentication Middleware**: JWT-based authentication
- **Admin Only Middleware**: Role-based authorization for admin routes
- **CORS Middleware**: Cross-Origin Resource Sharing policy enforcement with production-ready security defaults

### Error Handling System

The application implements a comprehensive error handling system:

- **Error Types**: Business errors, validation errors, and internal errors
- **Error Metadata**: Support for additional error context
- **Stack Traces**: Automatic stack trace capture for internal errors
- **Error Categorization**: Error classification with appropriate HTTP status codes
- **Structured Logging**: Detailed error logging with consistent format

### Task System Architecture

The application includes a robust task system with three types of task execution models:

- **Scheduled Tasks**: CRON-based tasks executed at specified intervals (e.g., daily reports, cleanup jobs)
- **Polling Tasks**: Tasks that run at fixed intervals to continuously check for conditions or data changes
- **Queue Tasks**: Asynchronous tasks processed via RabbitMQ for background processing and workload distribution

Each task type follows a consistent registration and execution pattern, making it easy to add new tasks while ensuring proper lifecycle management and error handling.

## Authentication System

The application implements a JWT-based authentication system with the following features:

- **Short-lived tokens**: By default, access tokens expire after 30 minutes for enhanced security
- **Token refresh**: Supports token refresh within a configurable time window (default 7 days)
- **Stateless design**: No server-side session storage, perfect for horizontal scaling
- **RESTful implementation**: Token passed via Authorization header

Frontend applications can integrate with the authentication system through the endpoints:

- `POST /api/web/v1/login` - For user authentication
- `GET /api/web/v1/refresh_token` - For refreshing expired tokens

## API Response Format

All API responses follow a consistent structure:

```json
{
  "code": 200,           // HTTP status code
  "message": "Success",  // Human-readable message
  "data": {},            // Response payload (when successful)
  "meta": {}             // Additional metadata (e.g., pagination)
}
```

Error responses maintain the same structure:

```json
{
  "code": 400,                 // Error code
  "message": "Validation error", // Error message
  "data": null,                // No data on error
  "meta": null                 // No metadata on error
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

3. Set up environment variables (use samples in config/dev as a starting point)

4. Run database migrations

   ```bash
   # Simplest approach with default options
   make db

   # Alternative commands
   make migrate

   # Or using the CLI directly with more options
   go run cmd/migrate/main.go up
   go run cmd/migrate/main.go -dir=db/migrations -verbose status
   ```

5. Start the API server

   ```bash
   go run cmd/web-api/main.go
   ```

6. Start the Task runner with desired components (all optional flags)

   ```bash
   go run cmd/task/main.go --scheduler --poller --queue
   ```

### Docker Deployment

1. Build the Docker image

   ```bash
   docker build -t go-layout .
   ```

2. Run with Docker Compose

   ```bash
   docker-compose up -d
   ```

### Production Deployment

For detailed production deployment instructions, refer to our [Deployment Guide](docs/deployment.md).

The project includes configuration for deploying to both single-node and multi-node k3s clusters:

```bash
# Deploy to a single node k3s environment
make deploy-single

# Deploy to a k3s cluster environment
make deploy-cluster

# Deploy to a specific server (defined in deploy/config/servers.yaml)
make deploy-server SERVER=servername
```

## API Endpoints

The API provides the following endpoints:

- **Authentication**
  - `POST /api/web/v1/login` - Login and get JWT token
  - `GET /api/web/v1/refresh_token` - Refresh JWT token
  - `GET /api/web/v1/captcha` - Get captcha for login

- **User Management**
  - `GET /api/web/v1/profile` - Get current user profile (requires authentication)
  - `PUT /api/web/v1/profile` - Update current user profile (requires authentication)
  - `POST /api/web/v1/users` - Create new user (requires admin role)
  - `GET /api/web/v1/users/:id` - Get user by ID (requires admin role)
  - `PUT /api/web/v1/users/:id` - Update user (requires admin role)
  - `PATCH /api/web/v1/users/:id/enabled` - Update user's enabled status (requires admin role)
  - `DELETE /api/web/v1/users/:id` - Delete user (requires admin role)
  - `GET /api/web/v1/users` - List users with pagination and filtering (requires admin role)

- **System**
  - `GET /health` - Health check endpoint

## License

[MIT](LICENSE)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## Database Migrations

This project uses [Goose](https://github.com/pressly/goose) for database migrations. Migrations are written in SQL and stored in the `db/migrations` directory.

### Migration Versioning Strategy

We use Goose's hybrid versioning approach to handle migrations in a team environment:

1. **Development**: Migrations are automatically timestamped (`20240628120000_add_users.sql`) during development, which helps avoid version conflicts when multiple developers create migrations simultaneously.

2. **Production**: Before deploying to production, migrations are converted to sequential versions while preserving their original order, which ensures predictable application in production.

This hybrid approach allows for flexible development while maintaining strict ordering in production.

### Running Migrations Locally

You can run database migrations using the following Makefile commands:

```bash
# Apply all pending migrations
make migrate

# Check migration status
make migrate-status

# Roll back the last migration
make migrate-down

# Roll back all migrations
make migrate-reset

# Create a new migration file (automatically timestamped)
make migrate-create NAME=add_users_table

# Fix migration versioning (converts timestamps to sequential numbers)
make migrate-fix

# Print current migration version
make migrate-version

# Rollback and reapply the latest migration
make migrate-redo

# Migrate up to a specific version
make migrate-up-to VERSION=20240628120000

# Migrate down to a specific version
make migrate-down-to VERSION=20240628120000
```

During local development, the system allows for out-of-order migrations, enabling team members to work independently without version conflicts. All migrations are validated through regex patterns to prevent SQL injection.

### Production Migrations

In production environments, migrations are handled by a Kubernetes Job. When deploying a new version, migrations should be run before deploying the application:

```bash
# Build the migration image (automatically fixes versioning)
make build-migrate

# Run migrations in a single node k3s environment
make deploy-migrate

# Run migrations in a k3s cluster environment
make deploy-migrate-cluster
```

The migration job runs as a Kubernetes job with `restartPolicy: Never` and `backoffLimit: 3`. It executes the migration operation once and then completes. You can check the job status and logs using:

```bash
kubectl get jobs -n go-layout
kubectl logs job/db-migrate -n go-layout
```

### CI/CD Integration

For CI/CD pipelines, use the `migrate-ci` command to prepare migrations for production:

```bash
# Convert timestamp versioning to sequential for production deployment
make migrate-ci
```

This should be run in your CI pipeline before building the migration image. The `build-migrate` command also automatically runs `migrate-fix` to ensure sequential versioning in production.

### Migration Files

Migration files follow the Goose naming convention:

```
{version}_{description}.sql
```

Each migration file contains two sections:
- `-- +goose Up`: Commands executed when migrating up
- `-- +goose Down`: Commands executed when rolling back the migration

Example migration file:

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

### Using Docker for Migrations

The project includes a dedicated `Dockerfile.migrate` for containerizing the migration tool. This container has an entrypoint configured to run the migration tool and defaults to showing the migration status. In Kubernetes jobs, we override this to run `up` command.

```bash
# Run a migration container locally (shows status by default)
docker run go-layout-migrate

# Run specific migration commands
docker run go-layout-migrate up
docker run go-layout-migrate down
```
