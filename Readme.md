# Go-Flow Workflow Engine 🚀

A production-ready sequential workflow engine with Go backend (Hexagonal Architecture) and Next.js frontend, featuring automatic retry logic, structured logging, and real-time monitoring.

## ✨ Features

- 🔄 **Sequential Task Execution** - Tasks execute in strict order with automatic orchestration
- ⚡ **Background Worker** - Concurrent task processing with configurable polling intervals
- 🔁 **Automatic Retry Logic** - Failed tasks retry with exponential backoff (2^n seconds)
- 📦 **Self-Contained Workflows** - Each workflow in its own package with clear organization
- 🏛️ **Clean Architecture** - Hexagonal Architecture (Ports & Adapters) for maintainability
- 🎯 **Type-Safe Database** - Jet ORM v2 generates type-safe queries from schema
- 📊 **Real-time Monitoring** - SWR auto-refresh every 1 second for live updates
- 🌙 **Dark Mode** - Persistent theme with smooth transitions

## 🏗️ Architecture

```
internal/
├── adapters/
│   ├── driven/        # Database Repository (Secondary Adapter)
│   └── driving/       # HTTP Handler (Primary Adapter)
├── core/
│   ├── domain/        # Domain Models
│   ├── port/          # Interface Definitions (Ports)
│   ├── registry/      # Workflow Registry (manages workflow definitions)
│   ├── service/       # Business Logic
│   └── worker/        # Background Worker
└── workflows/         # Self-Contained Workflow Packages
    ├── order/         # Order workflow with tasks
    └── refund/        # Refund workflow with tasks
```

**Hexagonal Architecture Components:**
- **Ports**: Interfaces defined in `port/` (WorkflowRepository, WorkflowService)
- **Adapters**: 
  - **Driving** (Primary): HTTP Handler receives external requests
  - **Driven** (Secondary): Database Repository connects to MySQL
- **Core**: Business logic and domain models independent of infrastructure

## 🛠️ Tech Stack

### Backend
- **Go** 1.25.3
- **Echo** v4 - High-performance web framework (migrated from Fiber)
- **Jet** v2 - Type-safe SQL Builder/ORM
- **MySQL** 8.0+ - Database
- **zerolog** - Structured logging
- **go-playground/validator** v10 - Input validation
- **godotenv** - Environment configuration
- **UUID** - Unique ID generation
- **Air** - Hot reload for development

### Frontend
- **Next.js** 16 - React framework with App Router
- **TypeScript** - Type safety
- **Tailwind CSS** v4 - Utility-first CSS (new syntax)
- **SWR** - Data fetching and caching
- **Axios** - HTTP client
- **Lucide React** - Icon library

## 📋 Prerequisites

- Go 1.25.3+
- MySQL 8.0+
- Node.js 18+ (for frontend)
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
    status ENUM('PENDING', 'RUNNING', 'COMPLETED', 'FAILED', 'RETRYING') DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    input_payload JSON,
    output_payload JSON,
    error_message TEXT,
    scheduled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_instance_id) REFERENCES workflow_instances(id)
);

CREATE TABLE activity_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    workflow_instance_id VARCHAR(36) NOT NULL,
    task_name VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    details JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_instance_id) REFERENCES workflow_instances(id)
);
```

### 3. Configure Environment

Create `.env` file in the root directory:

```env
# Application Environment
ENVIRONMENT=development  # or production

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

# Worker Configuration
WORKER_POLL_INTERVAL=5s
WORKER_BATCH_SIZE=10
WORKER_TASK_TIMEOUT=30s
WORKER_MAX_RETRIES=3
```

### 4. Install Dependencies

**Backend:**
```bash
go mod download
```

**Frontend:**
```bash
cd frontend
npm install  # or yarn install
```

### 5. Run Application

**Option 1: With Hot Reload (Recommended for Development)**

Backend:
```bash
air  # Automatically reloads on file changes
```

Frontend:
```bash
cd frontend
npm run dev  # or yarn dev
```

**Option 2: Manual Run**

Backend:
```bash
go run cmd/main.go
```

Frontend:
```bash
cd frontend
npm run build && npm start  # or yarn build && yarn start
```

**Access Points:**
- Backend API: `http://localhost:8080`
- Frontend UI: `http://localhost:3000`
- Health Check: `http://localhost:8080/health`
- Readiness Check: `http://localhost:8080/readiness`

## 📖 Usage

### List Available Workflows

Get all registered workflows:

```bash
curl http://localhost:8080/workflows/available
```

**Response:**
```json
{
  "workflows": ["OrderProcess", "RefundProcess"]
}
```

### Create Workflow Instance

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

## 📝 Define Custom Workflows

### Self-Contained Workflow Pattern

Go-Flow uses a **Self-Contained Workflow Pattern** where each workflow is in its own package with all related tasks. This provides:
- ✅ Clear separation of concerns
- ✅ No naming conflicts between workflows
- ✅ Easy to test and maintain
- ✅ Team can work on different workflows independently

### Step 1: Create Workflow Package

Create a new directory under `internal/workflows/`:

```bash
mkdir -p internal/workflows/user
```

### Step 2: Define Workflow Registration

Create `internal/workflows/user/workflow.go`:

```go
package user

import (
	"github.com/parinyadagon/go-workflow/internal/core/registry"
)

// Register registers the UserOnboarding workflow
func Register(reg *registry.WorkflowRegistry) {
	reg.NewWorkflow("UserOnboarding").
		AddTask("CreateAccount", createAccount).
		AddTask("SendWelcomeEmail", sendWelcomeEmail).
		AddTask("AssignRole", assignRole).
		MustBuild()
}
```

### Step 3: Implement Task Functions

Create `internal/workflows/user/tasks.go`:

```go
package user

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/parinyadagon/go-workflow/gen/go_flow/model"
	"github.com/parinyadagon/go-workflow/pkg/logger"
)

func createAccount(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "CreateAccount").Msg("Creating user account")

	// Parse input payload
	var input map[string]interface{}
	if task.InputPayload != nil {
		if err := json.Unmarshal([]byte(*task.InputPayload), &input); err != nil {
			return err
		}
	}

	// Business logic here
	email, ok := input["email"].(string)
	if !ok || email == "" {
		return errors.New("email is required")
	}

	// Simulate work
	time.Sleep(1 * time.Second)

	// Set output payload for next task
	output := map[string]interface{}{
		"user_id":    "USR-" + time.Now().Format("20060102150405"),
		"email":      email,
		"created_at": time.Now().Format(time.RFC3339),
	}
	outputJSON, _ := json.Marshal(output)
	outputStr := string(outputJSON)
	task.OutputPayload = &outputStr

	return nil
}

func sendWelcomeEmail(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "SendWelcomeEmail").Msg("Sending welcome email")

	// Parse input from previous task
	var input map[string]interface{}
	if task.InputPayload != nil {
		json.Unmarshal([]byte(*task.InputPayload), &input)
	}

	// Get data from previous task
	userID := input["user_id"].(string)
	email := input["email"].(string)

	// Send email logic here
	time.Sleep(500 * time.Millisecond)

	// Pass data to next task
	output := map[string]interface{}{
		"user_id":     userID,
		"email":       email,
		"email_sent":  true,
		"sent_at":     time.Now().Format(time.RFC3339),
	}
	outputJSON, _ := json.Marshal(output)
	outputStr := string(outputJSON)
	task.OutputPayload = &outputStr

	return nil
}

func assignRole(ctx context.Context, task *model.Tasks) error {
	logger.Info().Str("task", "AssignRole").Msg("Assigning default role")

	var input map[string]interface{}
	if task.InputPayload != nil {
		json.Unmarshal([]byte(*task.InputPayload), &input)
	}

	userID := input["user_id"].(string)

	// Assign role logic here
	time.Sleep(300 * time.Millisecond)

	output := map[string]interface{}{
		"user_id":     userID,
		"role":        "member",
		"assigned_at": time.Now().Format(time.RFC3339),
	}
	outputJSON, _ := json.Marshal(output)
	outputStr := string(outputJSON)
	task.OutputPayload = &outputStr

	return nil
}
```

### Step 4: Register Workflow in Main

Edit `cmd/main.go` and add your workflow:

```go
import (
	// ... other imports
	"github.com/parinyadagon/go-workflow/internal/workflows/order"
	"github.com/parinyadagon/go-workflow/internal/workflows/refund"
	"github.com/parinyadagon/go-workflow/internal/workflows/user"  // Add this
)

func main() {
	// ... setup code

	// Create registry and register workflows
	workflowRegistry := registry.NewWorkflowRegistry()
	
	order.Register(workflowRegistry)
	refund.Register(workflowRegistry)
	user.Register(workflowRegistry)  // Add this line

	// ... rest of the code
}
```

### Step 5: Test Your Workflow

```bash
curl -X POST http://localhost:8080/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "UserOnboarding",
    "input_payload": {
      "email": "user@example.com",
      "username": "newuser"
    }
  }'
```

## 🔗 Task Data Flow

Tasks communicate by passing data through `InputPayload` and `OutputPayload`:

```
Task 1: CreateAccount
  ├─ Input:  {"email": "user@example.com"}
  └─ Output: {"user_id": "USR-001", "email": "user@example.com"}
          ↓ (Worker passes Output as Input to next task)
Task 2: SendWelcomeEmail
  ├─ Input:  {"user_id": "USR-001", "email": "user@example.com"}
  └─ Output: {"user_id": "USR-001", "email_sent": true}
          ↓
Task 3: AssignRole
  ├─ Input:  {"user_id": "USR-001", "email_sent": true}
  └─ Output: {"user_id": "USR-001", "role": "member"}
```

**Key Points:**
- Each task reads from `task.InputPayload` (JSON string)
- Each task writes to `task.OutputPayload` (JSON string)
- Worker automatically passes `OutputPayload` of current task as `InputPayload` of next task
- Data persists throughout the workflow chain

## 🎨 Workflow Organization Patterns

### Option 1: Simple Tasks (Recommended for small workflows)

All tasks in one file:

```
internal/workflows/
└── simple/
    ├── workflow.go  # Register workflow
    └── tasks.go     # All task functions
```

### Option 2: Separate Task Files (Recommended for medium workflows)

Each task in its own file:

```
internal/workflows/
└── order/
    ├── workflow.go      # Register workflow
    ├── validate.go      # ValidateOrder task
    ├── payment.go       # DeductMoney task
    └── notification.go  # SendEmail task
```

### Option 3: Feature-Based (Recommended for complex workflows)

Group related tasks:

```
internal/workflows/
└── ecommerce/
    ├── workflow.go       # Register workflow
    ├── validation/       # Validation tasks
    │   ├── order.go
    │   └── inventory.go
    ├── payment/          # Payment tasks
    │   ├── authorize.go
    │   └── capture.go
    └── fulfillment/      # Fulfillment tasks
        ├── ship.go
        └── notify.go
```

## 🔄 Workflow Execution Flow

1. **Client** sends POST request to `/workflows`
2. **HTTP Handler** receives request, validates input, and calls Service
3. **Workflow Service** creates Workflow Instance and first Task (status: PENDING)
4. **Background Worker** (polls every 5 seconds, configurable):
   - Fetches Tasks with status = PENDING (batch size: 10, configurable)
   - Executes Tasks concurrently with Goroutines (with WaitGroup for safety)
   - **Retry Logic**:
     - If Task fails: checks retry count vs max retries (default: 3)
     - Updates status to RETRYING
     - Applies exponential backoff: 2^retryCount seconds (2s, 4s, 8s...)
     - Logs retry attempts in activity_logs
     - Marks as FAILED if max retries exceeded
   - When Task completes → creates next Task
   - When last Task completes → updates Workflow status = COMPLETED
5. **Activity Logs** track all events:
   - `TASK_STARTED` - Task execution begins
   - `TASK_RETRY` - Task retry attempt (with backoff delay)
   - `TASK_FAILED` - Task failed after max retries
   - `TASK_COMPLETED` - Task successfully completed
   - `WORKFLOW_COMPLETED` - Entire workflow finished

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/core/service/...
```

## 🎨 Frontend Features

### UI Components
- **Workflow List** - Paginated table with status badges
- **Workflow Detail** - Visual task pipeline with real-time updates
- **Status Indicators**:
  - ✅ COMPLETED - Green badge with checkmark
  - 🔄 RUNNING - Yellow badge with spinning animation
  - ❌ FAILED - Red badge with alert icon
  - ⏳ PENDING - Gray badge with circle icon

### User Experience
- **Dark Mode** - Full dark mode support with theme persistence (localStorage)
- **Real-time Updates** - SWR polling every 1 second
- **Error Boundary** - Graceful error handling with fallback UI
- **Loading States** - Skeleton loaders and spinners
- **Responsive Design** - Mobile-friendly layout

### Pagination
- Previous/Next navigation
- Page info: "Page X of Y"
- Total count: "Total: X workflows"
- Server-side pagination with metadata

### Tailwind CSS v4
- No config file required
- New `@import "tailwindcss"` syntax
- CSS variables for theming
- Dark mode variant: `@variant dark (&:where(.dark, .dark *))`

## 🔧 Development Tools

### Air - Hot Reload
```bash
# Install Air
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

### Jet - Code Generation
```bash
# Generate models from database
jet -dsn="mysql://user:pass@localhost:3306/go_flow" -path=./gen
```

## 📦 Project Structure

```
go-flow/
├── cmd/
│   └── main.go                    # Application entry point
├── config/
│   └── config.go                  # Configuration management (env vars)
├── db/
│   └── db.go                      # Database connection
├── pkg/
│   └── logger/
│       └── logger.go              # Structured logging (zerolog)
├── gen/                           # Generated code from Jet
│   └── go_flow/
│       ├── enum/                  # Enum types (statuses)
│       ├── model/                 # Database models
│       └── table/                 # Table definitions
├── internal/
│   ├── adapters/
│   │   ├── driven/
│   │   │   └── workflow_repo.go  # MySQL Repository
│   │   └── driving/
│   │       └── http_handler.go   # HTTP Handler (Echo)
│   ├── core/
│   │   ├── domain/                # Domain models
│   │   ├── port/
│   │   │   └── workflow.go       # Interfaces (Ports)
│   │   ├── registry/
│   │   │   └── workflow_builder.go # Workflow Registry
│   │   ├── service/
│   │   │   └── workflow_service.go # Business logic
│   │   └── worker/
│   │       └── workflow_worker.go  # Background worker (retry logic)
│   └── workflows/                 # Self-Contained Workflows
│       ├── order/                 # Order workflow
│       │   ├── workflow.go       # Register workflow
│       │   ├── validate.go       # ValidateOrder task
│       │   ├── payment.go        # DeductMoney task
│       │   └── notification.go   # SendEmail task
│       └── refund/                # Refund workflow
│           ├── workflow.go       # Register workflow
│           └── tasks.go          # All refund tasks
├── frontend/                      # Next.js Frontend
│   ├── app/
│   │   ├── components/            # React components
│   │   ├── workflows/[id]/        # Workflow detail page
│   │   ├── layout.tsx             # Root layout (theme provider)
│   │   ├── page.tsx               # Workflow list (with pagination)
│   │   └── globals.css            # Tailwind CSS v4
│   ├── package.json
│   └── tsconfig.json
├── .env                           # Environment variables
├── .env.example                   # Environment template
├── .air.toml                      # Air hot reload config
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
- `FAILED` - Execution failed (after max retries)

### Task
Sub-jobs in each workflow step:
- Each Workflow contains multiple Tasks
- Tasks execute sequentially
- Tasks are created when the previous Task completes
- **Retry Support**:
  - `retry_count`: Number of retry attempts (default: 0)
  - `RETRYING` status during retry attempts
  - Exponential backoff between retries
  - Configurable max retries (default: 3)

### Worker
Background process that:
- Polls Tasks with status = PENDING every 5 seconds (configurable)
- Processes Tasks concurrently with Goroutines
- Manages Workflow orchestration
- Handles retry logic with exponential backoff
- Uses WaitGroup to prevent race conditions
- Configurable: poll interval, batch size, task timeout, max retries

### Activity Logs
Complete audit trail of workflow execution:
- Tracks all workflow and task events
- Includes detailed JSON payloads
- Event types: TASK_STARTED, TASK_RETRY, TASK_FAILED, TASK_COMPLETED, WORKFLOW_COMPLETED
- Useful for debugging and monitoring

## 🚀 Production Features

### Structured Logging
- **zerolog** for high-performance structured logging
- Log levels: Debug, Info, Warn, Error, Fatal
- Development mode: Pretty console output with colors
- Production mode: JSON format for log aggregation
- Contextual fields: task_id, workflow_id, retry_count, etc.

### Error Handling
- Comprehensive error handling throughout the codebase
- No ignored errors (all `_` replaced with proper checks)
- Validation errors return field-level details
- Database errors logged with full context

### Health Checks
- `/health` - Basic liveness check (returns status: ok)
- `/readiness` - Readiness check with database ping
- Kubernetes-ready for liveness and readiness probes

### Input Validation
- go-playground/validator v10 for request validation
- Validation rules: required, min, max, email, etc.
- Returns structured validation errors with field names

### Concurrency Safety
- WaitGroup ensures all goroutines complete
- Proper context handling for cancellation
- No race conditions in worker processing

### Configuration
All settings via environment variables:
- Database connection pooling
- Worker behavior (poll interval, batch size, timeout)
- Retry logic (max retries)
- Server configuration

## 🔌 API Endpoints

| Method | Endpoint | Description | Query Params |
|--------|----------|-------------|--------------||
| GET | `/workflows/available` | List all registered workflows | - |
| POST | `/workflows` | Create a new Workflow | - |
| GET | `/workflows` | List all workflows with pagination | `limit`, `offset` |
| GET | `/workflows/:id` | Get workflow details with tasks and logs | - |
| GET | `/health` | Health check endpoint | - |
| GET | `/readiness` | Readiness check (includes DB ping) | - |

### API Examples

**List Workflows with Pagination:**
```bash
curl "http://localhost:8080/workflows?limit=10&offset=0"
```

**Response:**
```json
{
  "workflows": [...],
  "limit": 10,
  "offset": 0,
  "total": 45,
  "totalPages": 5,
  "currentPage": 1
}
```

**Get Workflow Details:**
```bash
curl "http://localhost:8080/workflows/550e8400-e29b-41d4-a716-446655440000"
```

**Response:**
```json
{
  "workflow": {...},
  "tasks": [...],
  "activityLogs": [
    {
      "event_type": "TASK_STARTED",
      "details": "{\"task_id\":1,\"task_name\":\"ValidateOrder\"}"
    },
    {
      "event_type": "TASK_RETRY",
      "details": "{\"retry_count\":1,\"backoff_delay\":\"2s\"}"
    },
    {
      "event_type": "TASK_COMPLETED",
      "details": "{\"status\":\"success\",\"retry_count\":1}"
    }
  ]
}
```

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

### Development Workflow
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `go test ./...`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Style
- Follow Go best practices and idioms
- Use `gofmt` for code formatting
- Write descriptive commit messages
- Add comments for complex logic
- Update README for new features

## 🗺️ Roadmap

- [ ] Unit tests for worker, service, and handlers
- [ ] Integration tests with test database
- [ ] Metrics and monitoring (Prometheus)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Workflow scheduling (cron support)
- [ ] Task dependencies (parallel execution)
- [ ] Workflow versioning
- [ ] Admin dashboard for workflow management
- [ ] Webhook notifications
- [ ] GraphQL API

## 📝 License

This project is licensed under the MIT License.

## 👤 Author

**Parinya Dagon**
- GitHub: [@parinyadagon](https://github.com/parinyadagon)

## 🙏 Acknowledgments

- [Echo](https://echo.labstack.com/) - High-performance web framework
- [Jet](https://github.com/go-jet/jet) - Type-safe SQL builder
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- [Next.js](https://nextjs.org/) - React framework
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS

---

⭐ If you like this project, please give it a star!