# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Laravel-style database migration system with tracking
- Comprehensive test suite (119+ tests)
- Migration documentation in `internal/database/MIGRATIONS.md`
- Complete open source documentation (README, CONTRIBUTING, LICENSE)
- GitHub issue and PR templates

## [1.0.0] - 2025-12-14

### Features

* **ci:** use semantic-release for automated versioning ([5a34dac](https://github.com/pandeptwidyaop/central-logs/commit/5a34dac6f342d234f73961e1a201c129c376e3ba))
* **ui:** replace native confirm dialogs with custom ConfirmDialog component ([54eccd6](https://github.com/pandeptwidyaop/central-logs/commit/54eccd69fd65eac9ea4fb1da5d71e3a4c9ce4576))

### Added
- Initial release
- Multi-project log aggregation
- Real-time log streaming via WebSocket
- JWT authentication with RBAC
- API key authentication for log ingestion
- Batch log ingestion endpoint
- Telegram notification integration
- Discord webhook integration
- Generic webhook support
- Web Push notifications (VAPID)
- Two-factor authentication (2FA)
- User management system
- Project permission system
- Advanced log filtering and search
- Log retention policies
- Rate limiting
- Statistics dashboard
- Beautiful React UI with Radix components
- Embedded frontend in single binary
- Docker support
- Comprehensive API documentation
- Development mode with hot reload

### Backend Stack
- Go 1.24+
- Fiber web framework
- SQLite database with WAL mode
- Redis for rate limiting (optional)
- JWT for authentication
- bcrypt for password hashing

### Frontend Stack
- React 19
- TypeScript
- Vite
- Tailwind CSS
- Radix UI
- React Router v7
- Lucide icons

### Developer Experience
- Make-based build system
- Comprehensive test coverage
- Hot reload for backend and frontend
- Database migrations with rollback
- ESLint + Prettier for frontend
- golangci-lint for backend

[Unreleased]: https://github.com/pandeptwidyaop/central-logs/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/pandeptwidyaop/central-logs/releases/tag/v1.0.0
