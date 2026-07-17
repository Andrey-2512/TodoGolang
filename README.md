<div align="center">

# 📋 Go TODO API

**High-performance Task Management API built with Go & Clean Architecture**

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-pgxpool-336791?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-cache-DC382D?style=flat-square&logo=redis&logoColor=white)](https://redis.io/)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](#)

</div>

---

## 📖 Table of Contents

- [Architecture Overview](#-architecture-overview)
- [Prerequisites](#-prerequisites)
- [API Reference](#-api-reference)
    - [Auth](#auth)
    - [Tasks](#tasks)
    - [Session](#session)
- [Configuration](#%EF%B8%8F-configuration)
- [Getting Started](#-getting-started)
- [Tech Stack](#-tech-stack)

---

## 🏗 Architecture Overview

The project follows a **layered (Clean) architecture**, keeping business logic isolated from infrastructure concerns:

| Layer | Responsibility |
|---|---|
| **Domain** | Entities and business logic interfaces |
| **Service** | Business logic implementation |
| **Repository** | PostgreSQL interaction via `pgxpool` |
| **Delivery** | HTTP handlers and middleware |

```
cmd/            → application entrypoint
internal/
  domain/       → entities & interfaces
  service/      → business logic
  repository/   → data access (Postgres)
  delivery/     → HTTP handlers, middleware, routing
pkg/
  database/     → migrations & DB connection setup
```

---

## ⚙️ Prerequisites

| Dependency | Version |
|---|---|
| Go | `1.22+` |
| PostgreSQL | `latest` |
| Redis | `latest` |

---

## 🔝 API Reference

> 🔒 = requires `Authorization` header (JWT)

### Auth

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/register/` | Register a new user |
| `POST` | `/login/` | Log in, returns JWT |
| `POST` | `/refresh/` 🔒 | Refresh JWT |
| `POST` | `/logout/` | Log out, returns `204 No Content` |

**Register / Login payload:**
```json
{
  "username": "EXAMPLE",
  "password": "EXAMPLE"
}
```

### Tasks

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/tasks/` 🔒 | Get all tasks |
| `GET` | `/tasks/{id}/` 🔒 | Get task by ID |
| `POST` | `/tasks/` 🔒 | Create a task |
| `PUT` | `/tasks/{id}` 🔒 | Full update of a task |
| `PATCH` | `/tasks/{id}` 🔒 | Partial update of a task |
| `DELETE` | `/tasks/{id}` 🔒 | Delete a task, returns `204 No Content` |

**Create / Full update payload:**
```json
{
  "title": "EXAMPLE",       // required
  "description": "EXAMPLE"  // optional
}
```

**Partial update (`PATCH`) payload:**
```json
{
  // send only the fields you want to update — at least one is required
  "title": "Update EXAMPLE",
  "description": "Update EXAMPLE"
}
```

### Session

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/me/` 🔒 | Current session info: task count & max. limit |

---

## ⚙️ Configuration

Create a `config.yaml` and `.env` file in the project root before running the app.

<details>
<summary><strong>config.yaml</strong></summary>

```yaml
hash:
  memory: 65536
  time: 3
  threads: 1
  key_len: 32
  salt_length: 16

http:
  addr: "0.0.0.0:8000"
  cors_url:
    - "http://localhost:5050"
  allow_credentials: true
  allow_methods:
    - "GET"
    - "POST"
    - "PATCH"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Authorization"
    - "Content-Type"
  access_control_max_age: 3800
  idle_timeout: "60s"
  read_timeout: "20s"
  write_timeout: "30s"
  cookie_secure: false
  handler_timeout: "5s"

jwt:
  access_ttl: "30m"
  refresh_ttl: "168h"
  whitelist_prefix: "wl:"

database:
  host: "localhost"
  port: 5432
  name: "go_todo_db"
  max_idle_conns: 20
  max_open_conns: 80
  max_conn_lifetime: "1h"
  conn_timeout: "10s"

redis:
  addr: "localhost:6379"
  db: 0
  min_idle_conns: 100
  pool_size: 100
  read_timeout: "500ms"
  write_timeout: "500ms"
  conn_max_lifetime: "1h"
  conn_timeout: "10s"

cache:
  cache_task_ttl: "30m"
  tasks_cache_prefix: "tasks:"
  user_tasks_cache_prefix: "user:"

app:
  max_tasks_per_user: 500
```

</details>

<details>
<summary><strong>.env</strong></summary>

```dotenv
JWT_SECRET_KEY="YOUR-SUPER-SECRET-KEY"
CONFIG_PATH="./config.yaml"     # path to your config.yaml
DB_USERNAME="YOUR-DB-USERNAME"
DB_PASSWORD="YOUR-DB-PASSWORD"
REDIS_USERNAME="YOUR-REDIS-USERNAME"
REDIS_PASSWORD="YOUR-REDIS-PASSWORD"
```

</details>

---

## 🏃 Getting Started

### 1. Run migrations

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -path pkg/database/postgres/migrations \
  -database "postgres://YOUR_POSTGRES_USERNAME:YOUR_POSTGRES_PASSWORD@localhost:5432/YOUR_DB_NAME?sslmode=disable" \
  up
```

### 2. Build & run

**Linux / macOS**

```bash
go build -o main ./cmd/main.go
./main
```

**Windows**

```bash
go build -o main.exe ./cmd/main.go
./main.exe
```

### ✅ Done!

The API should now be available at `http://localhost:8000` (or whatever `http.addr` you configured).

---

## 🧰 Tech Stack

- **Go 1.22+** — core language
- **pgxpool** — PostgreSQL driver & connection pooling
- **Redis** — caching layer
- **JWT** — stateless authentication
- **golang-migrate** — database migrations

