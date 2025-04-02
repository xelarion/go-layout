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
- **Middleware**: Modular middleware components for request processing, error handling, and request timeouts
- **Task Management**: Integrated support for scheduled tasks, polling tasks, and queue-based asynchronous processing

## Technology Stack

- **Programming Language**: Go
- **Web Framework**: Gin
- **ORM**: Gorm
- **Database**: PostgreSQL (configurable to switch to MySQL)
- **Cache**: Redis (pluggable, can be removed if not needed)
- **Message Queue**: RabbitMQ (pluggable, can be removed if not needed)
- **Logging**: Structured logging with zap
- **Configuration**: Environment-based configuration with github.com/caarlos0/env/v11
- **Error Handling**: Custom error package with metadata, stack traces, and error categorization
- **Scheduler**: Integrated CRON-based task scheduler using robfig/cron
- **Deployment**: Docker and k3s for containerization and orchestration

## Project Structure

```go
.
├── cmd/                   # Application entry points
│   ├── api/               # API server
│   └── task/              # Task runner for scheduled, polling and queue tasks
├── config/                # Configuration files
│   ├── dev/               # Development environment configs
│   └── prod/              # Production environment configs
├── db/                    # Database related files
│   └── migrations/        # Database migration files
├── internal/              # Private application code
│   ├── api/               # API-specific code
│   │   ├── public/        # Public API handlers and routers
│   │   └── web/           # Web API handlers and routers
│   │       ├── handler/   # API request handlers
│   │       ├── middleware/# HTTP middleware components
│   │       ├── types/     # Request/response structures
│   │       └── router.go  # Route definitions
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
   go run cmd/migrate/main.go
   ```

5. Start the API server

   ```bash
   go run cmd/api/main.go
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

## API Endpoints

The API provides the following endpoints:

- **Authentication**
  - `POST /api/v1/login` - Login and get JWT token
  - `GET /api/v1/refresh_token` - Refresh JWT token
  - `GET /api/v1/captcha` - Get captcha for login

- **User Management**
  - `GET /api/v1/profile` - Get current user profile (requires authentication)
  - `PUT /api/v1/profile` - Update current user profile (requires authentication)
  - `POST /api/v1/users` - Create new user (requires admin role)
  - `GET /api/v1/users/:id` - Get user by ID (requires admin role)
  - `PUT /api/v1/users/:id` - Update user (requires admin role)
  - `PATCH /api/v1/users/:id/enabled` - Update user's enabled status (requires admin role)
  - `DELETE /api/v1/users/:id` - Delete user (requires admin role)
  - `GET /api/v1/users` - List users with pagination and filtering (requires admin role)

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
