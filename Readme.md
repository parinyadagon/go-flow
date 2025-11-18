# Go-Flow Workflow Engine 🚀

A simple and lightweight workflow engine written in Go for managing sequential tasks using Hexagonal Architecture pattern.

## ✨ Features

- ✅ **Sequential Task Execution** - Execute tasks in a defined sequential order
- ✅ **Background Worker** - Asynchronous task processing with worker pool
- ✅ **Workflow Definition** - Easy workflow definition using maps
- ✅ **Type-Safe Database** - Type-safe SQL queries with Jet ORM
- ✅ **Clean Architecture** - Hexagonal Architecture (Ports & Adapters)
- ✅ **RESTful API** - HTTP API powered by Fiber Framework

## 🏗️ Architecture

```
internal/
├── adapters/
│   ├── driven/        # Database Repository (Secondary Adapter)
│   └── driving/       # HTTP Handler (Primary Adapter)
├── core/
│   ├── domain/        # Domain Models
│   ├── port/          # Interface Definitions (Ports)
│   ├── service/       # Business Logic
│   └── worker/        # Background Worker
```

**Hexagonal Architecture Components:**
- **Ports**: Interfaces defined in `port/` (WorkflowRepository, WorkflowService)
- **Adapters**: 
  - **Driving** (Primary): HTTP Handler receives external requests
  - **Driven** (Secondary): Database Repository connects to MySQL
- **Core**: Business logic and domain models independent of infrastructure

## 🛠️ Tech Stack

- **Go** 1.25.3
- **Fiber** v2 - Web Framework
- **Jet** v2 - Type-safe SQL Builder/ORM
- **MySQL** - Database
- **UUID** - Unique ID Generation

## 📋 Prerequisites

- Go 1.25.3+
- MySQL 8.0+
- Git

## 🚀 Getting Started

### 1. Clone Repository

```bash
git clone https://github.com/parinyadagon/go-workflow.git
cd go-workflow
```

### 2. Setup Database

Create database and tables in MySQL:

```sql
CREATE DATABASE go_flow;

USE go_flow;

CREATE TABLE workflow_instances (
    id VARCHAR(36) PRIMARY KEY,
    workflow_name VARCHAR(255) NOT NULL,
    status ENUM('PENDING', 'RUNNING', 'COMPLETED', 'FAILED') DEFAULT 'PENDING',
    current_input JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    workflow_instance_id VARCHAR(36) NOT NULL,
    task_name VARCHAR(255) NOT NULL,
    status ENUM('PENDING', 'RUNNING', 'COMPLETED', 'FAILED') DEFAULT 'PENDING',
    input_payload JSON,
    output_payload JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_instance_id) REFERENCES workflow_instances(id)
);

CREATE TABLE activity_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    workflow_instance_id VARCHAR(36) NOT NULL,
    task_id INT,
    action VARCHAR(255) NOT NULL,
    details JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_instance_id) REFERENCES workflow_instances(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
```

### 3. Configure Environment

Create `.env` file:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=your_password
DB_NAME=go_flow
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_LIFETIME=5m

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run Application

```bash
go run cmd/main.go
```

Server will start at `http://localhost:8080`

## 📖 Usage

### Create Workflow

Start a new workflow via API:

```bash
curl -X POST http://localhost:8080/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "OrderProcess",
    "input_payload": {
      "order_id": "ORD-001",
      "amount": 1500
    }
  }'
```

**Response:**

```json
{
  "message": "Workflow stated successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "workflow_name": "OrderProcess",
    "status": "PENDING",
    "current_input": "{\"order_id\":\"ORD-001\",\"amount\":1500}",
    "created_at": "2025-11-18T10:00:00Z"
  }
}
```

### Define Custom Workflow

Edit in `internal/core/service/workflow_service.go`:

```go
var WorkflowDefinitions = map[string][]string{
    "OrderProcess": {"ValidateOrder", "DeductMoney", "SendEmail"},
    "UserOnboarding": {"CreateAccount", "SendWelcomeEmail", "AssignRole"},
}
```

## 🔄 Workflow Execution Flow

1. **Client** sends POST request to `/workflows`
2. **HTTP Handler** receives request and calls Service
3. **Workflow Service** creates Workflow Instance and first Task (status: PENDING)
4. **Background Worker** (runs every 5 seconds):
   - Fetches Tasks with status = PENDING (max 10 tasks)
   - Executes Tasks concurrently with Goroutines
   - When Task completes → creates next Task
   - When last Task completes → updates Workflow status = COMPLETED

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## 📦 Project Structure

```
go-flow/
├── cmd/
│   └── main.go                    # Application entry point
├── config/
│   └── config.go                  # Configuration management
├── db/
│   └── db.go                      # Database connection
├── gen/                           # Generated code from Jet
│   └── go_flow/
│       ├── enum/                  # Enum types
│       ├── model/                 # Database models
│       └── table/                 # Table definitions
├── internal/
│   ├── adapters/
│   │   ├── driven/
│   │   │   └── workflow_repo.go  # MySQL Repository
│   │   └── driving/
│   │       └── http_handler.go   # HTTP Handler
│   └── core/
│       ├── domain/                # Domain models
│       ├── port/
│       │   └── workflow.go       # Interfaces
│       ├── service/
│       │   └── workflow_service.go # Business logic
│       └── worker/
│           └── workflow_worker.go  # Background worker
├── go.mod
├── go.sum
└── README.md
```

## 🎯 Core Concepts

### Workflow Instance
A created and running workflow with various statuses:
- `PENDING` - Waiting for processing
- `RUNNING` - Currently executing
- `COMPLETED` - Successfully finished
- `FAILED` - Execution failed

### Task
Sub-jobs in each workflow step:
- Each Workflow contains multiple Tasks
- Tasks execute sequentially
- Tasks are created when the previous Task completes

### Worker
Background process that:
- Polls Tasks with status = PENDING every 5 seconds
- Processes Tasks concurrently
- Manages Workflow orchestration

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/workflows` | Create a new Workflow |

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

## 📝 License

This project is licensed under the MIT License.

## 👤 Author

**Parinya Dagon**
- GitHub: [@parinyadagon](https://github.com/parinyadagon)

---

⭐ If you like this project, please give it a star!