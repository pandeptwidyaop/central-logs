# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Central Logs is a centralized log aggregation platform with a Go/Fiber backend and React/TypeScript frontend. The backend serves an embedded SPA and provides REST APIs for log ingestion, project management, user authentication, and notification channels.

## Build & Development Commands

```bash
# Development
make install          # Install Go deps, Air hot-reload, npm packages
make dev              # Run backend (Air) + frontend (Vite) concurrently
make dev-backend      # Backend only with hot-reload
make dev-frontend     # Frontend only

# Testing
make test             # Run all Go tests: go test -v ./...
go test -v ./internal/models/...     # Test specific package
go test -v ./... -run TestUserRepository  # Run specific test

# Building
make build            # Full production build (embeds frontend into Go binary)
make build-linux      # Cross-compile for Linux x86_64
make frontend         # Build React app only
make backend          # Build Go binary only

# Utilities
make clean            # Remove build artifacts
make db-reset         # Delete SQLite database
make docker-up        # Start Redis container
```

Frontend-specific commands (run from `frontend/`):
```bash
npm run dev           # Vite dev server on :5173
npm run build         # TypeScript check + Vite build
npm run lint          # ESLint
npx playwright test   # E2E tests
```

## Architecture

### Backend (Go/Fiber)

Request flow: Global Middleware → Auth/Validation → Handlers → Models (Repository pattern)

```
cmd/server/main.go     # Entry point, route registration
internal/
  config/              # YAML + env var configuration
  database/            # SQLite init and migrations
  handlers/            # HTTP handlers (auth, logs, projects, users, channels, stats)
  middleware/          # JWT auth, API key validation, RBAC, rate limiting
  models/              # Database models and repositories
  queue/               # Redis client for rate limiting
  utils/               # JWT manager
```

### Frontend (React 19 + Vite)

```
frontend/src/
  components/ui/       # Radix UI component library
  components/layout/   # App layout wrapper
  contexts/            # React Context (auth-context)
  hooks/               # Custom hooks
  lib/api.ts           # API client with token management
  pages/               # Route pages (dashboard, projects, logs, users, settings)
```

### Embedded SPA

Production build embeds the React app via `web/embed.go`. The Go binary serves the SPA at `/` and API at `/api`.

## API Authentication

Two authentication mechanisms:
1. **JWT** - For admin/user APIs. Token from `/api/auth/login`, sent as `Authorization: Bearer <token>`
2. **API Key** - For log ingestion. Project-specific key sent as `X-API-Key` header

## Key API Routes

```
# Auth
POST /api/auth/login
GET  /api/auth/me

# Log Ingestion (API Key auth)
POST /api/v1/logs
POST /api/v1/logs/batch

# Admin (JWT auth)
GET/POST   /api/admin/projects
GET/PUT/DELETE /api/admin/projects/:id
POST /api/admin/projects/:id/rotate-key
GET  /api/admin/logs
GET  /api/admin/stats/overview
```

## Configuration

Config loads from `config.yaml` with env var overrides using `CL_` prefix:
- `CL_SERVER_PORT=8080` overrides `server.port`
- `CL_JWT_SECRET=xxx` overrides `jwt.secret`

Default admin credentials are in config.yaml (change in production).

## Database

SQLite with WAL mode. Key tables: users, projects, user_projects (RBAC), logs, channels.

Tests use in-memory SQLite (`:memory:`) for isolation.

## Development Notes

- Backend hot-reload via Air watches `.go` files
- Frontend proxies `/api` and `/ws` to `localhost:3000` in dev mode
- Redis required for rate limiting features (optional in dev)
- Test files use `_test` package suffix for black-box testing
- gunakan playwright untuk e2e testing
- selalu gunakan npm run dev untuk FE agar live reload
- selalu gunakan air untuk hot reload backend