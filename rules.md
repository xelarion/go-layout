### 1. Project Background

- **Project Name**: go-layout
- **Objective**: A template project built with Go aimed at developing a high-availability, high-performance, and scalable web application.
- **Core Technologies**:
  - **Programming Language**: Go
  - **Web Framework**: Gin
  - **ORM**: Gorm
  - **Database**: PostgreSQL – designed for both horizontal and vertical scaling, with the possibility to switch to MySQL when necessary.
  - **Cache**: Redis – implemented as a pluggable component (to be removed if not needed), using the library [github.com/redis/go-redis/v9](http://github.com/redis/go-redis/v9).
  - **Message Queue**: RabbitMQ – also pluggable (to be removed if not needed), using the library [github.com/wagslane/go-rabbitmq](http://github.com/wagslane/go-rabbitmq).
  - **Deployment**: Containerized using Docker, deployed on a k3s cluster.

---

### 2. Core Requirements

- **High Availability**:
  - Adopt a stateless design to support cluster deployment.
  - Ensure stability under high concurrency and fault conditions.

- **High Performance**:
  - Optimize database queries to reduce latency and resource usage.
  - Apply appropriate caching strategies and conduct performance profiling and benchmarking on critical components.

- **Scalability**:
  - Use a modular design to support both horizontal (multiple modules) and vertical scaling.
  - Provide clear interfaces for replacing third-party components (e.g., Redis, RabbitMQ, JWT).

- **Clean and Elegant Code**:
  - Maintain clear, organized code adhering to best practices and coding standards.
  - Ensure the project is maintainable and extensible.

---

### 3. Technical Architecture & Modules

- **System Architecture**:
  - Stateless design
  - Cluster deployment strategy

- **Modular Design**:
  - **Web API**: Main interface for external requests.
  - **App API** (if applicable): Shares business logic with the Web API but may have different request and response parameters.
  - **Task Modules**: Includes scheduled tasks, polling tasks, and queue-based tasks.

- **Typical Call Chain**:
  - **Handler → Service → Usecase → Repository**

- **Layer Responsibilities**:
  - **Handler Layer**:
    - Responsible for receiving and parsing external requests.
    - Invokes the Service layer after processing the incoming data.
    - Must use dedicated request and response structures for every API endpoint.
  - **Service Layer**:
    - Acts as a bridge between the Handler and Usecase layers.
    - Handles data transformation and process coordination, without implementing the core business logic directly.
  - **Usecase Layer**:
    - Encapsulates the core business logic independently from the API layer.
    - Ensures that multiple APIs can reuse the same business rules.
    - Defines Repository interfaces following the dependency inversion principle.
  - **Repository Layer**:
    - Manages data access and interactions with the database.

- **API Response Structure**:
  - All API responses must follow a unified structure with these fields:
    - `code`: HTTP status code
    - `message`: Human-readable message
    - `data`: The response payload (when successful)
    - `meta`: Additional metadata

---

### 4. Configuration & Dependency Management

- **Configuration**:
  - Utilize environment files to manage configurations.
  - Parse configurations using [github.com/caarlos0/env/v11](https://github.com/caarlos0/env).

- **Database Migration**:
  - Manage database schema migrations using [github.com/golang-migrate/migrate/v4](https://github.com/golang-migrate/migrate).
  - Follow the workflow for creating new database tables:
    1. Create migration files for the new table
    2. Execute `make migrate` to create the database tables
    3. Run `make gen-model TABLE=table_name` to generate model structures in `internal/model/gen/`
    4. Create corresponding model files in `internal/model/` directory that embed the generated structures
    5. Proceed with implementing the business logic

- **Dependency Management**:
  - Use Go Modules (`go.mod` and `go.sum`) for dependency management.
  - Regularly update dependencies to incorporate the latest security patches and new features.

---

### 5. Security Guidelines

- **User Input Validation**:
  - Rigorously validate and sanitize all user inputs to prevent common vulnerabilities such as SQL injection and XSS attacks.

- **Authentication & Authorization**:
  - Implement authentication using JWT, designed to be extensible to allow custom implementations if needed.

- **Data Protection**:
  - Ensure sensitive data is protected by encrypting data in transit and at rest.

---

### 6. Performance Optimization

- **Database Optimization**:
  - Optimize SQL queries to minimize delays and resource consumption.

- **Caching Strategy**:
  - Apply Redis caching where appropriate to improve response times.

- **Performance Testing**:
  - Conduct thorough performance analysis and benchmarking on critical code paths to identify and resolve bottlenecks.

---

### 7. Error Handling & Logging

- **Error Handling**:
  - Implement comprehensive and consistent error handling across the application.

- **Logging**:
  - Utilize structured logging with zap to effectively capture and monitor errors and system status.

---

### 8. Documentation & Code Style

- **Documentation**:
  - Maintain up-to-date documentation for all modules and components.
  - Provide clear setup, usage, and contribution guidelines.
  - Use English for code and comments, with README files provided in English (`README.md`).

- **Code Style & Comments**:
  - Ensure that all code and comments are written in English.
  - Follow principles of clean, modular, and well-documented code to ensure readability and ease of extension.

---

### 9. Constraints & Additional Guidelines

- **API and Function Usage**:
  - Avoid using deprecated APIs and functions to ensure forward compatibility.

- **Code Review Process**:
  - All code changes must be reviewed and approved via pull requests before merging.

- **Git Commit Guidelines**:
  - Follow the Conventional Commits specification (e.g., feat(xxx): add new feature, fix: resolve bug) with optional scopes and clear descriptions.

- **Code Changes Scope**:
  - Do not modify functionality or logic unrelated to the current change.

---

### Additional Recommendations

- **Monitoring & Alerting**:
  - Consider integrating monitoring systems (e.g., Prometheus, Grafana) for real-time performance and health monitoring.
  - Integrate third-party error monitoring platforms (e.g., Sentry) to promptly detect and respond to system anomalies.

- **Testing Coverage**:
  - Develop a comprehensive testing strategy including unit tests, integration tests, and end-to-end tests to ensure system stability as the project scales.

- **DocumentationGuidelines**
1. When making changes to project architecture, technologies, or best practices, always update README.md and translate to README-zh.md
2. Ensure rules.md reflects the current state of the project for future AI reference.
