# [2.0.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.6.0...v2.0.0) (2025-12-20)


* feat!: improve WebSocket authentication with header-based tokens ([018e702](https://github.com/pandeptwidyaop/central-logs/commit/018e70209b2dd1c7a42e8c75f3add0ba5f4a944e))


### BREAKING CHANGES

* WebSocket authentication method changed from query parameters to Sec-WebSocket-Protocol header for improved security.

Migration required for v1.x users:
- Frontend must use new WebSocket(url, ['token', jwtToken]) instead of query params
- Removed user_id from WebSocket URLs (no longer exposed in logs/URLs)
- Updated auth-context to expose JWT token for WebSocket connections
- See docs/MIGRATION-v2.0.0.md for complete migration guide

Features Added:
- WebSocket authentication via Sec-WebSocket-Protocol (RFC 6455 compliant)
- Realtime log updates on dashboard, logs page, and project detail pages
- MCP (Model Context Protocol) server integration for AI agents
- Security improvements: environment-based CORS, constant-time token comparison
- Security headers middleware (CSP, X-Frame-Options, etc.)
- Comprehensive documentation: migration guide and MCP usage guide

Security Improvements:
- No token exposure in URLs or server logs
- No user ID exposure in WebSocket connections
- Constant-time comparison for API keys and MCP tokens
- Environment-based CORS configuration for production
- Added security headers for XSS/clickjacking protection

Documentation:
- docs/MIGRATION-v2.0.0.md - Migration guide from v1.x to v2.0.0
- docs/MCP-USAGE-GUIDE.md - Complete MCP integration guide
- docs/security-improvements.md - Security enhancements documentation

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>

# [1.6.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.5.0...v1.6.0) (2025-12-19)


### Bug Fixes

* improve JSON metadata viewer with syntax highlighting and text wrapping ([a34aea8](https://github.com/pandeptwidyaop/central-logs/commit/a34aea8d3df3be478b957d8ccd3c79f9b83c79e6)), closes [#282c34](https://github.com/pandeptwidyaop/central-logs/issues/282c34)


### Features

* add JSON syntax highlighting to log metadata viewer ([1778656](https://github.com/pandeptwidyaop/central-logs/commit/1778656a9bd015832e0347f757ede991790c51b4))

# [1.5.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.4.0...v1.5.0) (2025-12-19)


### Bug Fixes

* handle undefined api_key_prefix in clipboard copy ([9216abc](https://github.com/pandeptwidyaop/central-logs/commit/9216abce440a3d986afad6ec4d912e94fedcfe88))


### Features

* improve API key copy and management UX ([f84bd9b](https://github.com/pandeptwidyaop/central-logs/commit/f84bd9b06aea6eae876c991abc000273812215fe))
* improve log detail dialog UX and fix text overflow ([556d5c4](https://github.com/pandeptwidyaop/central-logs/commit/556d5c4807bfc92559706689e6bde03dc1c11543))

# [1.4.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.3.2...v1.4.0) (2025-12-16)


### Features

* add Telegram notification integration ([6efe010](https://github.com/pandeptwidyaop/central-logs/commit/6efe010fe07cbb870a9e6955a221530850bbd466))

## [1.3.2](https://github.com/pandeptwidyaop/central-logs/compare/v1.3.1...v1.3.2) (2025-12-15)


### Bug Fixes

* **ui:** prevent double 'v' in sidebar version display ([09745c0](https://github.com/pandeptwidyaop/central-logs/commit/09745c0d77df20d4e3b79e424c32c2e5548084dd))
* **ui:** prevent double 'v' prefix in version display ([ce44a5b](https://github.com/pandeptwidyaop/central-logs/commit/ce44a5b64a34816e3d360a3d40b04d07f820ea0f))

## [1.3.1](https://github.com/pandeptwidyaop/central-logs/compare/v1.3.0...v1.3.1) (2025-12-15)


### Bug Fixes

* **docker-compose:** use IPv4 for healthcheck ([fe6ccc8](https://github.com/pandeptwidyaop/central-logs/commit/fe6ccc8cf973ad7f34aa89a0216f47b6440ab317))
* **docker:** use IPv4 for healthcheck instead of localhost ([afe6ad6](https://github.com/pandeptwidyaop/central-logs/commit/afe6ad63b18cbbd9d58538cad5fbd3e83782c6cb))

# [1.3.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.2.0...v1.3.0) (2025-12-14)


### Bug Fixes

* **docker:** add env_file config to load environment variables ([bb00b8b](https://github.com/pandeptwidyaop/central-logs/commit/bb00b8bf1610b41bd0dc0e74366bd7a347f8dba3))


### Features

* **config:** add dual-format environment variable support ([144bde7](https://github.com/pandeptwidyaop/central-logs/commit/144bde73ba0d59dc385afe71856866619d702be4))
* **server:** explicitly bind to 0.0.0.0 for Docker compatibility ([809bfd0](https://github.com/pandeptwidyaop/central-logs/commit/809bfd0772af1b4c6d4bd0caf923490150909f62))

# [1.2.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.1.0...v1.2.0) (2025-12-14)


### Features

* **ui:** add automatic update notification banner ([5e99d9e](https://github.com/pandeptwidyaop/central-logs/commit/5e99d9edad91e0ae0f86b246935d603286b8692e))

# [1.1.0](https://github.com/pandeptwidyaop/central-logs/compare/v1.0.0...v1.1.0) (2025-12-14)


### Bug Fixes

* **tests:** complete handler test schema updates ([9a46851](https://github.com/pandeptwidyaop/central-logs/commit/9a468514288303ace3300669feb8de519d880d8f))
* **tests:** update all test schemas and fix test failures ([d71dd9d](https://github.com/pandeptwidyaop/central-logs/commit/d71dd9dafab420bb379cdff7586a2a22f83c734c))


### Features

* add Laravel-style migrations and comprehensive open source documentation ([46f285d](https://github.com/pandeptwidyaop/central-logs/commit/46f285d042c7c146149f2bbeae40471683612bfd))

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
