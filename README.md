# 📊 Central Logs

<div align="center">

![Central Logs](https://img.shields.io/badge/version-1.5.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**Centralized Log Aggregation Platform**

A modern, real-time log aggregation system with beautiful UI, built with Go and React.

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [API](#-api-documentation) • [Contributing](#-contributing)

</div>

---

## 📖 Overview

Central Logs is a self-hosted log aggregation platform that helps you collect, monitor, and analyze logs from multiple applications in one central location. Built with performance and developer experience in mind.

### Why Central Logs?

- 🚀 **Blazing Fast** - Built with Go and SQLite for maximum performance
- 🎨 **Beautiful UI** - Modern React interface with real-time updates
- 🔐 **Secure** - JWT authentication with RBAC (Role-Based Access Control)
- 🤖 **AI-Ready** - Built-in MCP server for AI agent integration (Claude Desktop compatible)
- 📱 **Push Notifications** - Web Push API support for critical alerts
- 🔔 **Multi-Channel Alerts** - Telegram, Discord, and Webhook integrations
- 🎯 **Smart Filtering** - Advanced log filtering and search capabilities
- 📊 **Analytics** - Built-in statistics and log level distribution
- 🔄 **Real-time** - WebSocket support for live log streaming
- 🐳 **Easy Deploy** - Single binary deployment or Docker

## ✨ Features

### Core Features

- **Multi-Project Support** - Manage logs from multiple applications
- **Real-time Streaming** - Live log updates via WebSocket
- **Advanced Filtering** - Filter by level, project, date range, and search
- **Batch Ingestion** - Efficient bulk log ingestion endpoint
- **Log Retention** - Configurable retention policies per log level
- **API Key Authentication** - Secure log ingestion with project-specific API keys

### Notification System

- **Telegram Integration** - Send alerts to Telegram channels/groups
- **Discord Webhooks** - Post notifications to Discord channels
- **Generic Webhooks** - Custom webhook endpoints for any service
- **Web Push Notifications** - Browser push notifications (VAPID)
- **Configurable Thresholds** - Set minimum log levels per channel

### MCP Server (AI Integration)

- **Model Context Protocol** - Built-in MCP server for AI agent integration
- **7 Query Tools** - query_logs, get_log, list_projects, get_project, get_stats, search_logs, get_recent_logs
- **Token Management** - Secure token-based authentication with activity tracking
- **Project-Based Access** - Fine-grained permissions per token
- **Claude Desktop Ready** - Works seamlessly with Claude Desktop and other MCP clients
- **Activity Monitoring** - Track all AI agent interactions and API usage

See [MCP Documentation](docs/mcp.md) for detailed setup and usage instructions.

### User Management

- **Role-Based Access Control** - Admin and User roles
- **Project Permissions** - Fine-grained project access control
- **2FA Support** - Two-factor authentication (TOTP)
- **User Profiles** - Customizable user profiles with avatars

### Developer Experience

- **RESTful API** - Clean and well-documented REST API
- **Embedded Frontend** - Single binary contains both backend and UI
- **Hot Reload** - Development mode with auto-reload
- **Comprehensive Tests** - 119+ tests with 100% core coverage
- **Laravel-style Migrations** - Database versioning with rollback support

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- Node.js 20+ and npm (for frontend development)
- SQLite3 (included in most systems)
- Redis (optional, for rate limiting and real-time features)

### Installation

#### Option 1: From Source

```bash
# Clone the repository
git clone https://github.com/pandeptwidyaop/central-logs.git
cd central-logs

# Install dependencies
make install

# Run database migrations
make build

# Start the server
./bin/server
```

#### Option 2: Using Docker

```bash
# Pull the image
docker pull ghcr.io/pandeptwidyaop/central-logs:latest

# Run the container
docker run -d \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name central-logs \
  ghcr.io/pandeptwidyaop/central-logs:latest
```

#### Option 3: Development Mode

```bash
# Install all dependencies
make install

# Start development servers (backend + frontend)
make dev

# Backend: http://localhost:3000
# Frontend: http://localhost:5173
```

### First Steps

1. **Access the application**: Open http://localhost:3000
2. **Login with default admin**:
   - Username: `admin`
   - Password: `changeme123`
3. **Create a project**: Click "New Project" to create your first project
4. **Get API Key**: Copy the API key from project settings
5. **Start sending logs**: Use the API to send logs (see examples below)

## 📚 Documentation

### Configuration

Create a `config.yaml` file:

```yaml
# Server Configuration
server:
  port: 3000
  env: production

# Database
database:
  path: ./data/central-logs.db

# Redis (optional)
redis:
  url: redis://localhost:6379

# JWT Authentication
jwt:
  secret: your-secret-key-here
  expiry: 24h

# Initial Admin User
admin:
  username: admin
  password: your-secure-password

# Log Retention
retention:
  enabled: true
  default:
    max_age: 30d
    max_count: 100000
  cleanup:
    schedule: "0 2 * * *"  # Daily at 2 AM

# Rate Limiting
rate_limit:
  api:
    requests_per_minute: 1000
```

### Environment Variables

You can override config values with environment variables using `CL_` prefix:

```bash
export CL_SERVER_PORT=8080
export CL_JWT_SECRET=my-secret-key
export CL_DATABASE_PATH=/var/lib/central-logs/db.sqlite
```

### Sending Logs

#### Using cURL

```bash
# Single log
curl -X POST http://localhost:3000/api/v1/logs \
  -H "X-API-Key: your-project-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "level": "error",
    "message": "Database connection failed",
    "metadata": {
      "error": "connection timeout",
      "host": "db-server-1"
    }
  }'

# Batch logs
curl -X POST http://localhost:3000/api/v1/logs/batch \
  -H "X-API-Key: your-project-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      {"level": "info", "message": "Server started"},
      {"level": "error", "message": "Failed to connect"}
    ]
  }'
```

#### Using Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type LogEntry struct {
    Level    string                 `json:"level"`
    Message  string                 `json:"message"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func sendLog(apiKey, level, message string) error {
    log := LogEntry{
        Level:   level,
        Message: message,
        Metadata: map[string]interface{}{
            "host": "app-server-1",
        },
    }

    body, _ := json.Marshal(log)
    req, _ := http.NewRequest("POST", "http://localhost:3000/api/v1/logs", bytes.NewReader(body))
    req.Header.Set("X-API-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

#### Using Node.js

```javascript
const axios = require('axios');

async function sendLog(apiKey, level, message, metadata = {}) {
  try {
    await axios.post('http://localhost:3000/api/v1/logs', {
      level,
      message,
      metadata
    }, {
      headers: {
        'X-API-Key': apiKey,
        'Content-Type': 'application/json'
      }
    });
  } catch (error) {
    console.error('Failed to send log:', error);
  }
}

// Usage
sendLog('your-api-key', 'error', 'Payment processing failed', {
  orderId: '12345',
  amount: 99.99
});
```

#### Using Python

```python
import requests
import json

def send_log(api_key, level, message, metadata=None):
    url = "http://localhost:3000/api/v1/logs"
    headers = {
        "X-API-Key": api_key,
        "Content-Type": "application/json"
    }
    data = {
        "level": level,
        "message": message,
        "metadata": metadata or {}
    }

    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Usage
send_log(
    "your-api-key",
    "error",
    "Database query failed",
    {"query": "SELECT * FROM users", "duration_ms": 5000}
)
```

## 🔧 Development

### Project Structure

```
central-logs/
├── cmd/
│   ├── server/          # Main server entry point
│   └── loggen/          # Log generator utility
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database layer
│   │   └── migrations/  # Laravel-style migrations
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Auth, RBAC, Rate limiting
│   ├── models/          # Data models and repositories
│   ├── services/        # Business logic services
│   ├── utils/           # Utility functions
│   └── websocket/       # WebSocket hub
├── frontend/            # React frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── pages/       # Page components
│   │   ├── contexts/    # React contexts
│   │   └── lib/         # API client and utilities
├── tests/               # Integration tests
└── web/                 # Embedded frontend assets
```

### Building

```bash
# Build backend only
make backend

# Build frontend only
make frontend

# Full production build (embeds frontend into binary)
make build

# Cross-compile for Linux
make build-linux

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/handlers/...

# Run tests with coverage
go test -v -cover ./...

# Run integration tests
go test -v ./tests/...
```

### Database Migrations

Central Logs uses a Laravel-style migration system:

```bash
# Migrations run automatically on server start
./bin/server

# Check migration status
# (programmatically via Go code - see internal/database/MIGRATIONS.md)
```

Create a new migration:

```go
// internal/database/migrations/20240115120000_add_feature.go
package migrations

import "database/sql"

type AddFeature struct{}

func (m *AddFeature) Name() string {
    return "20240115120000_add_feature"
}

func (m *AddFeature) Up(tx *sql.Tx) error {
    _, err := tx.Exec("ALTER TABLE projects ADD COLUMN new_field TEXT")
    return err
}

func (m *AddFeature) Down(tx *sql.Tx) error {
    // Rollback logic
    return nil
}
```

See [internal/database/MIGRATIONS.md](internal/database/MIGRATIONS.md) for detailed migration documentation.

## 📡 API Documentation

### Authentication

Two authentication methods are supported:

1. **JWT Token** (for admin/user APIs)
   ```bash
   # Login to get token
   curl -X POST http://localhost:3000/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "password"}'

   # Use token in subsequent requests
   curl http://localhost:3000/api/admin/projects \
     -H "Authorization: Bearer your-jwt-token"
   ```

2. **API Key** (for log ingestion)
   ```bash
   curl -X POST http://localhost:3000/api/v1/logs \
     -H "X-API-Key: your-project-api-key" \
     -d '{"level": "info", "message": "Hello"}'
   ```

### API Endpoints

#### Authentication
- `POST /api/auth/login` - User login
- `GET /api/auth/me` - Get current user
- `PUT /api/auth/profile` - Update profile
- `POST /api/auth/change-password` - Change password

#### Projects (Admin)
- `GET /api/admin/projects` - List all projects
- `POST /api/admin/projects` - Create project
- `GET /api/admin/projects/:id` - Get project details
- `PUT /api/admin/projects/:id` - Update project
- `DELETE /api/admin/projects/:id` - Delete project
- `POST /api/admin/projects/:id/rotate-key` - Rotate API key

#### Logs
- `POST /api/v1/logs` - Create single log (API Key auth)
- `POST /api/v1/logs/batch` - Create batch logs (API Key auth)
- `GET /api/admin/logs` - List logs (JWT auth)
- `GET /api/admin/logs/:id` - Get log details (JWT auth)

#### Users (Admin)
- `GET /api/admin/users` - List users
- `POST /api/admin/users` - Create user
- `GET /api/admin/users/:id` - Get user
- `PUT /api/admin/users/:id` - Update user
- `DELETE /api/admin/users/:id` - Delete user

#### Statistics
- `GET /api/admin/stats/overview` - System overview stats

#### MCP Server (AI Integration)
- `POST /api/mcp/message` - MCP protocol endpoint
- `GET /api/admin/mcp/status` - Get MCP server status
- `POST /api/admin/mcp/toggle` - Enable/disable MCP server
- `GET /api/admin/mcp/tokens` - List MCP tokens
- `POST /api/admin/mcp/tokens` - Create MCP token
- `GET /api/admin/mcp/tokens/:id` - Get token details
- `PUT /api/admin/mcp/tokens/:id` - Update token
- `DELETE /api/admin/mcp/tokens/:id` - Delete token
- `GET /api/admin/mcp/tokens/:id/activity` - Get token activity logs

See [MCP Documentation](docs/mcp.md) for detailed API specifications and tool descriptions.

#### WebSocket
- `GET /ws` - WebSocket connection for real-time logs

### Log Levels

Supported log levels (in order of severity):

- `DEBUG` - Detailed debugging information
- `INFO` - Informational messages
- `WARN` - Warning messages
- `ERROR` - Error messages
- `CRITICAL` - Critical issues requiring immediate attention

## 🐳 Docker Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  central-logs:
    image: ghcr.io/pandeptwidyaop/central-logs:latest
    ports:
      - "3000:3000"
    volumes:
      - ./data:/data
      - ./config.yaml:/app/config.yaml
    environment:
      - CL_SERVER_PORT=3000
      - CL_JWT_SECRET=${JWT_SECRET}
    restart: unless-stopped
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped

volumes:
  redis-data:
```

Run with:

```bash
docker-compose up -d
```

## 🔒 Security

- All passwords are hashed using bcrypt
- JWT tokens for session management
- API keys are hashed before storage
- CORS protection enabled
- Rate limiting on all endpoints
- Input validation and sanitization
- SQL injection protection via prepared statements
- 2FA support for enhanced security

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Coding Standards

- Go code follows `gofmt` and `golint` standards
- React code follows ESLint configuration
- Write tests for new features
- Update documentation for API changes

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Fiber](https://gofiber.io/) - Express-inspired web framework
- UI components from [Radix UI](https://www.radix-ui.com/)
- Styled with [Tailwind CSS](https://tailwindcss.com/)
- Icons from [Lucide](https://lucide.dev/)

## 📮 Support

- 🐛 Issues: [GitHub Issues](https://github.com/pandeptwidyaop/central-logs/issues)
- 💡 Discussions: [GitHub Discussions](https://github.com/pandeptwidyaop/central-logs/discussions)

## 🗺️ Roadmap

- [x] **MCP Server Integration** - AI agent integration via Model Context Protocol ✅
- [ ] Elasticsearch integration
- [ ] Log parsing and structured logging
- [ ] Custom dashboards
- [ ] Alert rules engine
- [ ] Multi-tenant support
- [ ] S3/Object storage archiving
- [ ] Grafana integration
- [ ] Mobile app

---

<div align="center">

Made with ❤️ by the Central Logs Team

[⭐ Star us on GitHub](https://github.com/pandeptwidyaop/central-logs)

</div>
