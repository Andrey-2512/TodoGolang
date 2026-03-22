**High-Performance Task Management API built on Golang**

This TODO API on golang with Clean architecture

## Performance ⚡

This API shows **12 000 RPS** on benchmark on standard computer

Tested with `hey`
```text
Requests/sec: 12243.8
Average latency: 0.0081s
P99 latency: 0.0266s
```

## Architecture Overview 🏗
The project follows the layered architecture pattern:
- **Domain:** Entities and business logic interfaces.
- **Service:** Business logic implementation.
- **Repository:** PostgreSQL interaction using `pgxpool`.
- **Delivery:** HTTP handlers and middleware layer.


## Prerequisites
- Go 1.22+
- PostgreSQL
- Redis

## 1. Endpoints 🔝
```http request
 GET /tasks/ Return All Tasks
Authorization Required
 ```

```http request
GET /tasks/{id}/ Return Task By id
Authorization Required
```

```http request
POST /tasks/ Create Task Return Created Task

Arguments:
{
    "title": "EXAMPLE",
                             # required at least 1 field
    "description": "EXAMPLE"
}

Authorization Required
```

```http request
DELETE /tasks/{id} Delete Task Return 204 No Content

Authorization Required
```

```http request
PUT /tasks/ Dynamic Update Task Return Update Task

Arguments:
{
    # You can send 1 field which you wnat update but 1 required
    "title": "Update EXAMPLE",
                                    # required at least 1 field
    "description": "Update EXAMPLE"
}

Authorization Required
```

```http request
POST /register/ Return Registered User Username
Arguments:
{
    "username": "EXAMPLE", # required
    "password": "EXAMPLE   # required
}
```

```http request
POST /login/ Return Authorization Token JWT
Arguments:
{
    "username": "EXAMPLE", # required
    "password": "EXAMPLE   # required
}
```

```http request
POST /refresh/ Return Refreshed Authorization Token JWT

Authorization Required

```

```http request
POST /logout/ This endpoint is exiting your account Return 204 No Content
```

## 2. Configuration ⚙


You must create .yaml and .env files

**For example:**

**config.yaml**

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
  - "*"
  idle_timeout: "60s"
  read_timeout: "20s"
  write_timeout: "30s"

jwt:
  access_ttl: "30m"
  refresh_ttl: "168h"
  whitelist_prefix: "wl:"

database:
  host: "localhost"
  port: 5432
  name: "go_todo_db"
  max_idle_conns: 100
  max_open_conns: 100
  max_conn_lifetime: "1h"
  conn_timeout: "30s"

redis:
  addr: "localhost:6379"
  db: 0
  min_idle_conns: 100
  pool_size: 100
  read_timeout: "500ms"
  write_timeout: "500ms"
  conn_max_lifetime: "1h"

cache:
  cache_task_ttl: "30m"
  tasks_cache_prefix: "tasks:"
  user_tasks_cache_prefix: "user:"


```

**.env**
```dotenv
JWT_SECRET_KEY="YOUR-SUPER-SECRET-KEY"
CONFIG_PATH="./config.yaml" # Path to your config.yaml
DB_USERNAME="YOUR-DB-USERNAME"
DB_PASSWORD="YOUR-DB-PASSWORD"
REDIS_PASSWORD="YOUR-REDIS-PASSWORD"
```

## 3. Run 🏃‍♂️
**Next step open your console and run next commands:**

**Windows:**

```bash
go build PATH/TO/PROJECT/cmd/main.go

./main.exe
```

**Linux/MacOS:**

```Terminal
go build PATH/TO/PROJECT/cmd/main.go

./main
```


**Congratulations you can use this project!**
