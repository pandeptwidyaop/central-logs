# Central Logs - Project Planning

## Overview
Central Logs adalah tool untuk menerima dan mengelola logs dari berbagai aplikasi dengan notifikasi multi-channel. Mendukung multi-user dengan assignment per project.

## Tech Stack
- **Backend**: Go + Fiber v2 + **Go Embed** (single binary deployment)
- **Database**: SQLite
- **Queue/Rate Limiter**: Redis
- **Frontend**: React + Vite + Shadcn/ui + **PWA**
- **Notifications**: Web Push API, Telegram Bot, Discord Webhook

---

## Database Schema

### 1. users
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| email | TEXT | Unique email |
| password | TEXT | Hashed password (bcrypt) |
| name | TEXT | Display name |
| role | TEXT | ADMIN, USER |
| is_active | BOOLEAN | Status aktif |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### 2. projects
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| name | TEXT | Project name |
| description | TEXT | Project description |
| api_key | TEXT | Hashed API key untuk authentication |
| api_key_prefix | TEXT | Prefix API key (untuk display, misal: "cl_xxx...") |
| is_active | BOOLEAN | Status aktif project |
| created_at | DATETIME | Timestamp created |
| updated_at | DATETIME | Timestamp updated |

### 3. user_projects (pivot table)
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| user_id | TEXT | Foreign key ke users |
| project_id | TEXT | Foreign key ke projects |
| role | TEXT | OWNER, MEMBER, VIEWER |
| created_at | DATETIME | |

### 4. logs
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| project_id | TEXT | Foreign key ke projects |
| level | TEXT | DEBUG, INFO, WARN, ERROR, CRITICAL |
| message | TEXT | Log message |
| metadata | TEXT (JSON) | Additional data (stack trace, context, etc) |
| source | TEXT | Source/module dari log |
| timestamp | DATETIME | Timestamp dari client |
| created_at | DATETIME | Timestamp server received |

### 5. channels
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| project_id | TEXT | Foreign key ke projects |
| type | TEXT | PUSH, TELEGRAM, DISCORD |
| name | TEXT | Channel name |
| config | TEXT (JSON) | Channel-specific config |
| min_level | TEXT | Minimum level untuk notify (DEBUG/INFO/WARN/ERROR/CRITICAL) |
| is_active | BOOLEAN | Status aktif |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### 6. push_subscriptions
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| user_id | TEXT | Foreign key ke users |
| project_id | TEXT | Foreign key ke projects (nullable, null = all assigned projects) |
| endpoint | TEXT | Push subscription endpoint |
| p256dh | TEXT | Public key |
| auth | TEXT | Auth secret |
| user_agent | TEXT | Browser info |
| created_at | DATETIME | |

### 7. notification_history
| Column | Type | Description |
|--------|------|-------------|
| id | TEXT (UUID) | Primary key |
| log_id | TEXT | Foreign key ke logs |
| channel_id | TEXT | Foreign key ke channels |
| status | TEXT | PENDING, SENT, FAILED, RATE_LIMITED |
| error_message | TEXT | Error jika failed |
| sent_at | DATETIME | |

---

## User Roles & Permissions

### System Roles
| Role | Description |
|------|-------------|
| ADMIN | Full access: manage users, all projects, system settings |
| USER | Access only to assigned projects |

### Project Roles
| Role | Description |
|------|-------------|
| OWNER | Full control: edit project, manage channels, invite members, delete |
| MEMBER | View logs, manage channels |
| VIEWER | View logs only |

---

## API Endpoints

### Auth API
```
POST   /api/auth/login              - Login, returns JWT
POST   /api/auth/logout             - Logout (invalidate token)
GET    /api/auth/me                 - Get current user info
PUT    /api/auth/me                 - Update profile
PUT    /api/auth/change-password    - Change password
```

### Public API (untuk client apps)
```
POST /api/v1/logs
Header: X-API-Key: <project_api_key>
Body: {
  "level": "ERROR",
  "message": "Something went wrong",
  "metadata": { "user_id": 123, "stack": "..." },
  "source": "auth-service",
  "timestamp": "2024-01-15T10:30:00Z"  // optional
}
Response: { "id": "<log_id>", "status": "received" }
```

```
POST /api/v1/logs/batch
Header: X-API-Key: <project_api_key>
Body: {
  "logs": [
    { "level": "INFO", "message": "...", ... },
    { "level": "ERROR", "message": "...", ... }
  ]
}
```

### Admin API (untuk web UI - requires JWT)

#### Users (ADMIN only)
```
GET    /api/admin/users                 - List all users
POST   /api/admin/users                 - Create user
GET    /api/admin/users/:id             - Get user detail
PUT    /api/admin/users/:id             - Update user
DELETE /api/admin/users/:id             - Delete user
```

#### Projects
```
GET    /api/admin/projects              - List projects (filtered by user access)
POST   /api/admin/projects              - Create project (user becomes OWNER)
GET    /api/admin/projects/:id          - Get project detail
PUT    /api/admin/projects/:id          - Update project (OWNER/ADMIN only)
DELETE /api/admin/projects/:id          - Delete project (OWNER/ADMIN only)
POST   /api/admin/projects/:id/rotate-key - Rotate API key (OWNER/ADMIN only)
```

#### Project Members
```
GET    /api/admin/projects/:id/members      - List project members
POST   /api/admin/projects/:id/members      - Add member to project
PUT    /api/admin/projects/:id/members/:uid - Update member role
DELETE /api/admin/projects/:id/members/:uid - Remove member
```

#### Logs
```
GET    /api/admin/logs                  - List logs (filtered by user's projects)
GET    /api/admin/logs/:id              - Get log detail
DELETE /api/admin/logs                  - Bulk delete logs
```

#### Channels
```
GET    /api/admin/projects/:id/channels     - List channels
POST   /api/admin/projects/:id/channels     - Create channel
PUT    /api/admin/channels/:id              - Update channel
DELETE /api/admin/channels/:id              - Delete channel
POST   /api/admin/channels/:id/test         - Test channel
```

#### Push Notifications
```
POST   /api/admin/push/subscribe        - Subscribe to push
DELETE /api/admin/push/unsubscribe      - Unsubscribe
GET    /api/admin/push/vapid-public-key - Get VAPID public key
```

#### Stats
```
GET    /api/admin/stats/overview        - Dashboard stats (user's projects)
GET    /api/admin/stats/projects/:id    - Project-specific stats
```

#### WebSocket (Realtime Log Viewer)
```
WS /api/ws/logs?token=<jwt>&projects=<id1,id2>

// Connect with JWT token and optional project filter
// If projects param empty, subscribes to all user's assigned projects

// Client -> Server Messages:
{
  "type": "subscribe",
  "projects": ["project-id-1", "project-id-2"],  // Subscribe to specific projects
  "levels": ["ERROR", "CRITICAL"]                 // Optional: filter by levels
}

{
  "type": "unsubscribe",
  "projects": ["project-id-1"]
}

{
  "type": "ping"
}

// Server -> Client Messages:
{
  "type": "log",
  "data": {
    "id": "log-uuid",
    "project_id": "project-uuid",
    "project_name": "My Project",
    "level": "ERROR",
    "message": "Something went wrong",
    "metadata": { ... },
    "source": "auth-service",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}

{
  "type": "subscribed",
  "projects": ["project-id-1", "project-id-2"]
}

{
  "type": "error",
  "message": "Unauthorized access to project"
}

{
  "type": "pong"
}
```

---

## Architecture Flow

```
┌─────────────┐     POST /api/v1/logs     ┌─────────────────┐
│ Client App  │ ────────────────────────► │   Fiber API     │
└─────────────┘      + API Key            │   (Go Backend)  │
                                          └────────┬────────┘
                                                   │
          ┌────────────────────────────────────────┼────────────────────────────────────────┐
          │                    │                   │                   │                    │
          ▼                    ▼                   ▼                   ▼                    ▼
   ┌───────────────┐   ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐  ┌──────────────┐
   │    SQLite     │   │     Redis       │ │  Notification   │ │   WebSocket     │  │    Redis     │
   │  (Store Log)  │   │ (Rate Limiter)  │ │    Worker       │ │     Hub         │  │   Pub/Sub    │
   └───────────────┘   │    (Queue)      │ └────────┬────────┘ └────────┬────────┘  │  (Realtime)  │
                       └─────────────────┘          │                   │           └──────┬───────┘
                                          ┌─────────┼─────────┐         │                  │
                                          │         │         │         │                  │
                                          ▼         ▼         ▼         ▼                  │
                                   ┌──────────┐ ┌────────┐ ┌───────┐ ┌──────────┐         │
                                   │ Web Push │ │Telegram│ │Discord│ │ Browser  │◄────────┘
                                   └──────────┘ └────────┘ └───────┘ │  (Live)  │
                                                                     └──────────┘
```

---

## Folder Structure

```
central-logs/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── sqlite.go
│   │   └── migrations/
│   ├── models/
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── user_project.go
│   │   ├── log.go
│   │   ├── channel.go
│   │   └── subscription.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── users.go
│   │   ├── logs.go
│   │   ├── projects.go
│   │   ├── members.go
│   │   ├── channels.go
│   │   ├── push.go
│   │   └── websocket.go
│   ├── websocket/
│   │   ├── hub.go               # WebSocket connection hub
│   │   ├── client.go            # WebSocket client handler
│   │   └── message.go           # Message types
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication
│   │   ├── apikey.go            # API key auth for public endpoints
│   │   ├── rbac.go              # Role-based access control
│   │   └── ratelimit.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── log_service.go
│   │   ├── notification_service.go
│   │   └── channel/
│   │       ├── push.go
│   │       ├── telegram.go
│   │       └── discord.go
│   ├── queue/
│   │   └── redis.go
│   └── utils/
│       ├── apikey.go
│       ├── jwt.go
│       └── validator.go
│
├── web/                             # Embedded frontend
│   ├── embed.go                     # Go embed directive
│   └── dist/                        # Built frontend (generated)
│
├── frontend/                        # Frontend source
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/                  # Shadcn components
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── Header.tsx
│   │   │   │   └── AppShell.tsx
│   │   │   ├── projects/
│   │   │   ├── logs/
│   │   │   ├── channels/
│   │   │   └── users/
│   │   ├── pages/
│   │   │   ├── Login.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Projects.tsx
│   │   │   ├── ProjectDetail.tsx
│   │   │   ├── Logs.tsx
│   │   │   ├── Users.tsx            # Admin only
│   │   │   └── Settings.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useProjects.ts
│   │   │   ├── useLogs.ts
│   │   │   ├── usePushNotification.ts
│   │   │   └── useWebSocket.ts          # WebSocket hook for realtime logs
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx
│   │   ├── services/
│   │   │   └── api.ts
│   │   ├── lib/
│   │   │   ├── utils.ts
│   │   │   └── push-manager.ts      # Web Push utilities
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── public/
│   │   ├── sw.js                    # Service Worker (Push + PWA)
│   │   ├── manifest.json            # PWA Web App Manifest
│   │   └── icons/
│   │       ├── icon-192x192.png
│   │       ├── icon-512x512.png
│   │       └── apple-touch-icon.png
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
│
├── Makefile                         # Build commands
├── config.yaml                      # Application configuration
├── .air.toml                        # Air hot reload config
├── go.mod
├── go.sum
├── docker-compose.yml               # Redis + optional services
├── .env.example
└── README.md
```

---

## Go Embed Setup

### web/embed.go
```go
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// GetFileSystem returns the embedded frontend files
func GetFileSystem() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
```

### Serving Embedded Files (in main.go)
```go
import (
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"central-logs/web"
)

// Serve embedded frontend
frontendFS, _ := web.GetFileSystem()
app.Use("/", filesystem.New(filesystem.Config{
	Root:         http.FS(frontendFS),
	Browse:       false,
	Index:        "index.html",
	NotFoundFile: "index.html", // SPA fallback
}))
```

### Makefile
```makefile
.PHONY: dev dev-backend dev-frontend build clean frontend backend install

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Air for hot reload..."
	@go install github.com/air-verse/air@latest
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

# Development (run both backend and frontend with hot reload)
dev:
	@echo "Starting development servers..."
	@echo "Backend: http://localhost:3000 (API)"
	@echo "Frontend: http://localhost:5173 (Vite dev server)"
	@make -j2 dev-backend dev-frontend

dev-backend:
	@air -c .air.toml

dev-frontend:
	@cd frontend && npm run dev

# Build for production
build: frontend backend
	@echo "Build complete: ./bin/central-logs"

frontend:
	@echo "Building frontend..."
	@cd frontend && npm ci && npm run build
	@rm -rf web/dist
	@cp -r frontend/dist web/dist

backend:
	@echo "Building backend..."
	@CGO_ENABLED=1 go build -o bin/central-logs ./cmd/server

# Cross compile (Linux)
build-linux:
	@make frontend
	@echo "Building for Linux..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/central-logs-linux ./cmd/server

# Clean
clean:
	@rm -rf bin/ web/dist/ frontend/dist/ frontend/node_modules/ tmp/

# Run production
run:
	@./bin/central-logs
```

### .air.toml (Go Hot Reload Configuration)
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "web", "frontend", "node_modules"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
  include_file = ["config.yaml"]
  kill_delay = "2s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = true

[screen]
  clear_on_rebuild = true
  keep_scroll = true
```

### Development Commands
```bash
# First time setup
make install

# Start development (Air + Vite)
make dev

# This runs concurrently:
# - Backend (Air): http://localhost:3000 - auto-reload on .go/.yaml changes
# - Frontend (Vite): http://localhost:5173 - HMR for React

# Build for production
make build

# Run production binary
make run
# or
./bin/central-logs
```

### Frontend Vite Config for Development Proxy
```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
```

---

## PWA Configuration

### manifest.json
```json
{
  "name": "Central Logs",
  "short_name": "CentralLogs",
  "description": "Centralized logging and notification system",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0f172a",
  "theme_color": "#3b82f6",
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any maskable"
    },
    {
      "src": "/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any maskable"
    }
  ]
}
```

### Service Worker Features
- **Push Notifications**: Receive and display log notifications
- **Offline Support**: Basic offline page / cached assets
- **Background Sync**: Queue failed API requests when offline

---

## Implementation Phases

### Phase 1: Foundation
- [ ] Setup Go project dengan Fiber v2
- [ ] Setup SQLite database + migrations
- [ ] Setup Redis connection
- [ ] Implement user model + password hashing
- [ ] Implement JWT authentication
- [ ] Create initial admin user on first run

### Phase 2: User & Project Management
- [ ] Users CRUD (admin only)
- [ ] Projects CRUD
- [ ] User-Project assignment (pivot table)
- [ ] Role-based access control middleware
- [ ] API key authentication untuk public endpoints

### Phase 3: Core Log API
- [ ] POST /api/v1/logs endpoint
- [ ] POST /api/v1/logs/batch endpoint
- [ ] Log validation dan storage
- [ ] Basic rate limiting per project

### Phase 4: Admin API
- [ ] Logs list dengan filtering/pagination (respect user access)
- [ ] Channels CRUD endpoints
- [ ] Stats endpoints

### Phase 5: Notification System
- [ ] Redis queue untuk notifications
- [ ] Notification worker/consumer
- [ ] Web Push implementation (VAPID)
- [ ] Telegram Bot integration
- [ ] Discord Webhook integration
- [ ] Rate limiter per channel

### Phase 6: Frontend - Core
- [ ] Setup React + Vite + Shadcn
- [ ] PWA setup (manifest.json, service worker registration)
- [ ] Login page + Auth context
- [ ] Layout (Sidebar, Header) with role-based menu
- [ ] Dashboard dengan stats

### Phase 7: Frontend - Features
- [ ] Projects management page
- [ ] Project members management
- [ ] Logs viewer dengan filters
- [ ] Channels configuration
- [ ] Users management (admin only)
- [ ] Settings page + Push notification subscription UI

### Phase 8: Polish
- [ ] Service worker for push notifications
- [ ] Offline support
- [ ] Error handling improvement
- [ ] Docker setup
- [ ] Documentation

---

## Channel Config Examples

### Telegram
```json
{
  "bot_token": "123456:ABC-DEF...",
  "chat_id": "-1001234567890"
}
```

### Discord
```json
{
  "webhook_url": "https://discord.com/api/webhooks/..."
}
```

### Web Push (stored in push_subscriptions table)
```json
{
  "endpoint": "https://fcm.googleapis.com/fcm/send/...",
  "keys": {
    "p256dh": "...",
    "auth": "..."
  }
}
```

---

## Rate Limiting Strategy

### Per Channel Rate Limits (Redis)
- Key format: `ratelimit:channel:{channel_id}:{window}`
- Default limits:
  - Telegram: 20 messages/minute
  - Discord: 30 messages/minute
  - Web Push: 60 notifications/minute

### Per Project API Rate Limits
- Key format: `ratelimit:api:{project_id}:{window}`
- Default: 1000 requests/minute

---

## Log Level Priority
```
DEBUG    = 0  (lowest)
INFO     = 1
WARN     = 2
ERROR    = 3
CRITICAL = 4  (highest)
```

Channel dengan `min_level = ERROR` hanya akan menerima notifikasi untuk ERROR dan CRITICAL.

---

## Environment Variables

```env
# Server
PORT=3000
ENV=development

# Database
DATABASE_PATH=./data/central-logs.db

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRY=24h

# Web Push (VAPID)
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_SUBJECT=mailto:admin@example.com

# Initial Admin (created on first run)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=changeme123
```

---

## Configuration (config.yaml)

Aplikasi menggunakan `config.yaml` untuk konfigurasi yang lebih kompleks, termasuk log retention policy.

### config.yaml
```yaml
# Server Configuration
server:
  port: 3000
  env: development  # development, production

# Database
database:
  path: ./data/central-logs.db

# Redis
redis:
  url: redis://localhost:6379

# JWT Authentication
jwt:
  secret: your-super-secret-key
  expiry: 24h

# Web Push (VAPID)
vapid:
  public_key: ""
  private_key: ""
  subject: mailto:admin@example.com

# Initial Admin User (created on first run)
admin:
  email: admin@example.com
  password: changeme123

# Log Retention Configuration
retention:
  enabled: true

  # Global default retention (applies to all projects unless overridden)
  default:
    max_age: 30d              # Delete logs older than 30 days (format: 1d, 7d, 30d, 90d, 1y)
    max_count: 100000         # Max logs per project (0 = unlimited)

  # Per-level retention (optional, overrides default for specific levels)
  levels:
    debug:
      max_age: 7d             # DEBUG logs deleted after 7 days
      max_count: 10000
    info:
      max_age: 14d
      max_count: 50000
    warn:
      max_age: 30d
      max_count: 50000
    error:
      max_age: 90d            # Keep ERROR logs longer
      max_count: 100000
    critical:
      max_age: 365d           # Keep CRITICAL logs for 1 year
      max_count: 0            # Unlimited count for critical

  # Cleanup schedule
  cleanup:
    enabled: true
    schedule: "0 2 * * *"     # Cron format: Run at 2 AM daily
    batch_size: 1000          # Delete in batches to avoid lock

  # Notification history retention
  notification_history:
    max_age: 7d               # Keep notification history for 7 days

# Rate Limiting
rate_limit:
  # API rate limits
  api:
    requests_per_minute: 1000

  # Channel rate limits
  channels:
    telegram:
      messages_per_minute: 20
    discord:
      messages_per_minute: 30
    push:
      notifications_per_minute: 60
```

### Per-Project Retention Override

Projects dapat memiliki retention policy sendiri yang override global config.

#### Database Schema Addition (projects table)
| Column | Type | Description |
|--------|------|-------------|
| retention_config | TEXT (JSON) | Custom retention config (nullable, null = use global) |

#### Example Project Retention Config (stored in DB)
```json
{
  "max_age": "90d",
  "max_count": 500000,
  "levels": {
    "debug": { "max_age": "1d" },
    "critical": { "max_age": "2y" }
  }
}
```

### Retention Priority Order
1. **Per-Project Per-Level** - Highest priority
2. **Per-Project Default** - Project-wide setting
3. **Global Per-Level** - From config.yaml levels section
4. **Global Default** - From config.yaml default section

### Config Loading Priority
1. `config.yaml` (file)
2. Environment variables (override file, format: `CL_SERVER_PORT`, `CL_DATABASE_PATH`, etc.)
3. Command line flags (highest priority)

### Duration Format
Supported formats for `max_age`:
- `1d` - 1 day
- `7d` - 7 days
- `30d` - 30 days
- `90d` - 90 days
- `1y` / `365d` - 1 year
- `2y` / `730d` - 2 years

---

## Log Cleanup Worker

### How it Works
1. **Scheduler** runs cleanup job based on cron schedule
2. **Cleanup Process**:
   - Load retention config (global + per-project overrides)
   - For each project:
     - Delete logs exceeding `max_age` per level
     - Delete oldest logs if count exceeds `max_count` per level
   - Clean notification_history older than configured
3. **Batching**: Delete in batches to prevent database lock

### Manual Cleanup API (Admin Only)
```
POST /api/admin/maintenance/cleanup
Body: {
  "project_id": "optional-project-id",  // null = all projects
  "dry_run": true                       // Preview what would be deleted
}
Response: {
  "deleted_count": 1234,
  "by_level": {
    "debug": 500,
    "info": 400,
    "warn": 200,
    "error": 100,
    "critical": 34
  }
}
```

### Stats API Addition
```
GET /api/admin/stats/storage
Response: {
  "total_logs": 500000,
  "by_project": [
    { "project_id": "...", "name": "Project A", "count": 250000, "size_mb": 150 }
  ],
  "by_level": {
    "debug": 100000,
    "info": 200000,
    "warn": 100000,
    "error": 80000,
    "critical": 20000
  },
  "retention_status": {
    "last_cleanup": "2024-01-15T02:00:00Z",
    "next_cleanup": "2024-01-16T02:00:00Z",
    "last_deleted_count": 5000
  }
}
```

---

## WebSocket Realtime Log Viewer

### How it Works

```
┌─────────────┐  POST /api/v1/logs   ┌──────────────┐  Publish   ┌───────────────┐
│ Client App  │ ──────────────────►  │   API        │ ─────────► │ Redis Pub/Sub │
└─────────────┘                      │   Handler    │            │ (logs:project)│
                                     └──────────────┘            └───────┬───────┘
                                                                         │
                                                                         │ Subscribe
                                                                         ▼
┌─────────────┐  WebSocket           ┌──────────────┐            ┌───────────────┐
│  Browser    │ ◄──────────────────  │  WebSocket   │ ◄───────── │   Hub         │
│  (React)    │                      │   Client     │            │   Manager     │
└─────────────┘                      └──────────────┘            └───────────────┘
```

### Backend Implementation

#### Hub Manager (internal/websocket/hub.go)
```go
type Hub struct {
    // Registered clients by user ID
    clients    map[string]map[*Client]bool

    // Project subscriptions: projectID -> clients
    projects   map[string]map[*Client]bool

    // Channel for broadcasting logs
    broadcast  chan *LogMessage

    // Register/unregister channels
    register   chan *Client
    unregister chan *Client

    // Redis pub/sub for multi-instance support
    redisSub   *redis.PubSub
}

func (h *Hub) Run() {
    // Subscribe to Redis channel for all projects
    // Broadcast received messages to connected clients
}
```

#### Client Handler (internal/websocket/client.go)
```go
type Client struct {
    hub           *Hub
    conn          *websocket.Conn
    userID        string
    allowedProjects []string  // Projects user has access to
    subscribedProjects []string
    levelFilter   []string    // Optional level filter
    send          chan []byte
}

func (c *Client) ReadPump()  // Handle incoming messages
func (c *Client) WritePump() // Send messages to client
```

#### Log Ingestion Flow
```go
// In log handler, after saving to DB:
func (h *LogHandler) CreateLog(c *fiber.Ctx) error {
    // 1. Save log to SQLite
    log := saveLog(...)

    // 2. Publish to Redis for realtime distribution
    redis.Publish(ctx, "logs:"+log.ProjectID, logJSON)

    // 3. Queue notification if level matches channel config
    queueNotification(log)

    return c.JSON(log)
}
```

### Frontend Implementation

#### useWebSocket Hook
```typescript
// frontend/src/hooks/useWebSocket.ts
import { useEffect, useRef, useCallback, useState } from 'react'

interface LogMessage {
  id: string
  project_id: string
  project_name: string
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'CRITICAL'
  message: string
  metadata?: Record<string, any>
  source?: string
  timestamp: string
}

interface UseWebSocketOptions {
  projects?: string[]
  levels?: string[]
  onLog?: (log: LogMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
  autoReconnect?: boolean
}

export function useWebSocket(options: UseWebSocketOptions) {
  const [isConnected, setIsConnected] = useState(false)
  const [logs, setLogs] = useState<LogMessage[]>([])
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()

  const connect = useCallback(() => {
    const token = localStorage.getItem('token')
    const wsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/api/ws/logs?token=${token}`

    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      setIsConnected(true)
      options.onConnect?.()

      // Subscribe to projects
      if (options.projects?.length) {
        ws.send(JSON.stringify({
          type: 'subscribe',
          projects: options.projects,
          levels: options.levels
        }))
      }
    }

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data)

      if (message.type === 'log') {
        const log = message.data as LogMessage
        setLogs(prev => [log, ...prev].slice(0, 1000)) // Keep last 1000
        options.onLog?.(log)
      }
    }

    ws.onclose = () => {
      setIsConnected(false)
      options.onDisconnect?.()

      // Auto reconnect
      if (options.autoReconnect !== false) {
        reconnectTimeoutRef.current = setTimeout(connect, 3000)
      }
    }

    wsRef.current = ws
  }, [options])

  const subscribe = useCallback((projects: string[], levels?: string[]) => {
    wsRef.current?.send(JSON.stringify({
      type: 'subscribe',
      projects,
      levels
    }))
  }, [])

  const unsubscribe = useCallback((projects: string[]) => {
    wsRef.current?.send(JSON.stringify({
      type: 'unsubscribe',
      projects
    }))
  }, [])

  const disconnect = useCallback(() => {
    clearTimeout(reconnectTimeoutRef.current)
    wsRef.current?.close()
  }, [])

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [])

  return {
    isConnected,
    logs,
    subscribe,
    unsubscribe,
    disconnect,
    clearLogs: () => setLogs([])
  }
}
```

#### Realtime Log Viewer Component
```typescript
// frontend/src/components/logs/RealtimeLogViewer.tsx
import { useWebSocket } from '@/hooks/useWebSocket'
import { useState } from 'react'

export function RealtimeLogViewer({ projectIds }: { projectIds: string[] }) {
  const [paused, setPaused] = useState(false)
  const [levelFilter, setLevelFilter] = useState<string[]>([])

  const { isConnected, logs, clearLogs } = useWebSocket({
    projects: projectIds,
    levels: levelFilter.length ? levelFilter : undefined,
    autoReconnect: true
  })

  const filteredLogs = paused ? [] : logs

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2 p-2 border-b">
        <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
        <span className="text-sm">{isConnected ? 'Connected' : 'Disconnected'}</span>

        <button onClick={() => setPaused(!paused)}>
          {paused ? 'Resume' : 'Pause'}
        </button>

        <button onClick={clearLogs}>Clear</button>

        {/* Level filter buttons */}
        {['DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'].map(level => (
          <button
            key={level}
            className={levelFilter.includes(level) ? 'active' : ''}
            onClick={() => toggleLevel(level)}
          >
            {level}
          </button>
        ))}
      </div>

      {/* Log list */}
      <div className="flex-1 overflow-auto font-mono text-sm">
        {filteredLogs.map(log => (
          <LogEntry key={log.id} log={log} />
        ))}
      </div>
    </div>
  )
}
```

### Multi-Instance Support (Redis Pub/Sub)

Untuk mendukung multiple server instances, gunakan Redis Pub/Sub:

```go
// Publish log ke Redis (di log handler)
func publishLog(ctx context.Context, log *models.Log) {
    channel := fmt.Sprintf("logs:%s", log.ProjectID)
    data, _ := json.Marshal(log)
    redisClient.Publish(ctx, channel, data)
}

// Hub subscribe ke Redis (di hub manager)
func (h *Hub) subscribeRedis(ctx context.Context) {
    pubsub := h.redis.PSubscribe(ctx, "logs:*")

    for msg := range pubsub.Channel() {
        // Extract project ID from channel name
        projectID := strings.TrimPrefix(msg.Channel, "logs:")

        // Broadcast to all clients subscribed to this project
        h.broadcastToProject(projectID, msg.Payload)
    }
}
```

### Config Addition
```yaml
# config.yaml
websocket:
  enabled: true
  ping_interval: 30s      # Ping clients every 30s
  pong_timeout: 10s       # Wait 10s for pong response
  max_message_size: 512   # Max message size in KB

  # Buffer settings
  read_buffer_size: 1024
  write_buffer_size: 1024
```

---

## First Run Setup
Pada saat pertama kali server dijalankan:
1. Load config dari `config.yaml` (atau create default jika tidak ada)
2. Database migrations akan dijalankan otomatis
3. Jika belum ada user, akan dibuat admin user dengan credentials dari config
4. Start cleanup scheduler berdasarkan config
5. Initialize WebSocket hub dan Redis pub/sub subscription
6. Admin dapat login dan mulai membuat projects serta invite users
