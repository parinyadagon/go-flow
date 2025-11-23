# Go-Flow Workflow Engine - AI Coding Instructions

## Project Overview
Go-Flow is a sequential workflow engine with Go backend (Hexagonal Architecture) and Next.js frontend. Tasks execute sequentially via background workers polling every 5 seconds.

## Architecture Pattern: Hexagonal Architecture (Ports & Adapters)

### Key Concepts
- **Ports** (`internal/core/port/`): Interface definitions (e.g., `WorkflowRepository`, `WorkflowService`)
- **Adapters**:
  - **Driving/Primary** (`internal/adapters/driving/`): HTTP handlers (Echo framework)
  - **Driven/Secondary** (`internal/adapters/driven/`): Database repository (MySQL + Jet ORM)
- **Core** (`internal/core/`): Business logic, independent of infrastructure

### Data Flow
```
Client → Echo Handler → Service Layer → Repository → MySQL
                                      ↓
                                  Worker (polls tasks every 5s)
```

## Backend (Go)

### Framework & Dependencies
- **Web Framework**: Echo v4 (migrated from Fiber - PascalCase JSON responses)
- **ORM**: Jet v2 (type-safe SQL, generates code in `gen/go_flow/`)
- **Database**: MySQL 8.0+
- **Dev Tool**: Air (hot reload, excludes `frontend/` dir)

### Critical Conventions

#### 1. JSON Field Names: PascalCase (Important!)
Go structs use PascalCase field names without JSON tags. Frontend expects PascalCase:
```go
// Backend returns: {"ID": "...", "WorkflowName": "...", "Status": "..."}
// Frontend interface must match:
interface Workflow {
  ID: string;           // NOT id or workflow_id
  WorkflowName: string; // NOT workflowName
  Status: string;
}
```

#### 2. Database Connection
- Config: `.env` file (loaded via `godotenv`)
- Connection string format: `username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local`
- Use `config.Load()` then `db.NewConnection(&cfg.Database)`

#### 3. Workflow Definitions
Defined in `internal/core/service/workflow_service.go`:
```go
var WorkflowDefinitions = map[string][]string{
    "OrderProcess": {"ValidateOrder", "DeductMoney", "SendEmail"},
}
```
To add workflows, update this map.

#### 4. Worker Orchestration
- Background goroutine in `internal/core/worker/workflow_worker.go`
- Polls `GetTaskPending()` every 5 seconds (via ticker)
- Processes tasks concurrently with goroutines
- After task completion: creates next task OR marks workflow COMPLETED
- Started in `main.go`: `go workerNode.Start(ctx)`

#### 5. Status Enums
Use generated enums from `gen/go_flow/enum/`:
- `workflow_instances_status.go`: PENDING, RUNNING, COMPLETED, FAILED
- `tasks_status.go`: Same statuses

### Development Workflow

#### Run Backend
```bash
# With hot reload (recommended):
air

# Or manual:
go run cmd/main.go
```

Server runs on `http://localhost:8080`

#### API Endpoints
- `POST /workflows` - Create workflow
- `GET /workflows` - List workflows (with limit/offset pagination)
- `GET /workflows/:id` - Get workflow details with tasks

#### CORS Configuration
Frontend (`http://localhost:3000`) is whitelisted in `cmd/main.go`:
```go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"http://localhost:3000"},
}))
```

## Frontend (Next.js + TypeScript)

### Stack
- **Framework**: Next.js 16 (App Router)
- **Styling**: Tailwind CSS v4 (new syntax, no config file)
- **Data Fetching**: SWR (auto-revalidation)
- **HTTP Client**: Axios
- **Icons**: Lucide React
- **Theme**: Context API with localStorage

### Critical Conventions

#### 1. Tailwind CSS v4 Syntax Differences
```css
/* globals.css - No tailwind.config.js! */
@import "tailwindcss";

/* Dark mode variant syntax: */
@variant dark (&:where(.dark, .dark *));

/* NO gradient classes (bg-gradient-to-r) - use solid colors */
/* Use CSS variables for theme colors */
```

#### 2. Dark Mode Implementation
- ThemeProvider uses `documentElement.classList.add('dark')` (not data attribute)
- Must handle hydration: render placeholder until mounted
- Theme persisted in localStorage
- Example pattern in `app/components/ThemeProvider.tsx` and `ThemeToggle.tsx`

#### 3. Real-time Updates with SWR
```tsx
const { data, error, mutate } = useSWR<WorkflowData>(
  `http://localhost:8080/workflows/${id}`,
  fetcher,
  { refreshInterval: 1000 } // Poll every 1 second
);

// Manual refresh (don't use window.location.reload):
const handleRefresh = async () => {
  setIsRefreshing(true);
  await mutate();
  setTimeout(() => setIsRefreshing(false), 500);
};
```

#### 4. Component Structure
- `app/layout.tsx`: Root layout with ThemeProvider + Navbar
- `app/page.tsx`: Workflow list (table view)
- `app/workflows/[id]/page.tsx`: Workflow detail (task visualization)
- `app/components/`: Shared components (Navbar, ThemeToggle, ThemeProvider)

#### 5. Status Visualization
Task status styling in `workflows/[id]/page.tsx`:
- COMPLETED: Green border, checkmark icon
- RUNNING: Yellow border, spinning clock icon, animate-pulse
- FAILED: Red border, alert icon
- PENDING: Gray border, circle icon

#### 6. Loading & Error States
All data-fetching pages must handle:
```tsx
if (error) return <ErrorCard />; // With retry button using mutate()
if (!data) return <LoadingSpinner />; // Animated spinner
```

### Development Workflow

#### Run Frontend
```bash
cd frontend
npm run dev  # or yarn dev
```

Runs on `http://localhost:3000`

#### File Organization
```
app/
├── layout.tsx           # Root layout (ThemeProvider + Navbar)
├── page.tsx             # Workflow list
├── workflows/[id]/
│   └── page.tsx         # Workflow detail
├── components/
│   ├── Navbar.tsx       # Sticky top nav (must be opaque, not transparent)
│   ├── ThemeProvider.tsx
│   └── ThemeToggle.tsx
└── globals.css          # Tailwind v4 imports
```

## Common Patterns

### Backend: Adding New Endpoint
1. Define interface in `internal/core/port/workflow.go`
2. Implement in repository (`internal/adapters/driven/workflow_repo.go`)
3. Add service method (`internal/core/service/workflow_service.go`)
4. Create handler (`internal/adapters/driving/http_handler.go`)
5. Register route in `cmd/main.go`

### Frontend: Adding New Page
1. Create route file: `app/your-route/page.tsx`
2. Add "use client" directive if using hooks
3. Import types matching backend PascalCase
4. Use SWR for data fetching
5. Add loading/error states
6. Link from Navbar if needed

### Theme-aware Styling
```tsx
className="bg-white dark:bg-slate-900 text-gray-900 dark:text-white"
```

## Debugging Tips

### Backend Issues
- Check Air logs: `cat build-errors.log`
- Verify `.env` file exists and has correct DB credentials
- Test DB connection: `mysql -u root -p go_flow`

### Frontend Issues
- Check browser console for hydration errors
- Verify API responses match TypeScript interfaces (PascalCase!)
- Test API directly: `curl http://localhost:8080/workflows`

### Common Pitfalls
1. ❌ Using `fiber` imports (migrated to Echo)
2. ❌ Frontend expecting snake_case (backend returns PascalCase)
3. ❌ Tailwind v4 gradient classes (not supported)
4. ❌ Transparent navbar causing overlap
5. ❌ `window.location.reload()` instead of SWR `mutate()`

## Testing Strategy
Currently no automated tests. Manual testing via:
- Backend: `curl` commands
- Frontend: Browser + React DevTools
- Worker: Watch database for task status changes

## User Language
Project owner prefers communication in Thai (ภาษาไทย) but code/docs in English.
