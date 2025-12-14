# Contributing to Central Logs

First off, thank you for considering contributing to Central Logs! It's people like you that make Central Logs such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title** for the issue
- **Describe the exact steps which reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed after following the steps**
- **Explain which behavior you expected to see instead and why**
- **Include screenshots and animated GIFs** if possible
- **Include your environment details** (OS, Go version, Node version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate the steps**
- **Describe the current behavior and explain which behavior you expected to see instead**
- **Explain why this enhancement would be useful**

### Pull Requests

- Fill in the required template
- Do not include issue numbers in the PR title
- Follow the Go and JavaScript/TypeScript style guides
- Include thoughtfully-worded, well-structured tests
- Document new code based on the Documentation Styleguide
- End all files with a newline

## Development Setup

### Prerequisites

```bash
# Install Go 1.24+
# Install Node.js 20+
# Install SQLite3
# Install Redis (optional)
```

### Setup

1. Fork the repo
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/central-logs.git
   cd central-logs
   ```

3. Install dependencies:
   ```bash
   make install
   ```

4. Create a feature branch:
   ```bash
   git checkout -b feature/my-new-feature
   ```

5. Start development:
   ```bash
   make dev
   ```

### Project Structure

```
central-logs/
├── cmd/                 # Application entry points
├── internal/            # Private application code
│   ├── config/         # Configuration
│   ├── database/       # Database layer
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   └── services/       # Business logic
├── frontend/            # React frontend
├── tests/              # Integration tests
└── web/                # Embedded assets
```

## Development Workflow

### Backend Development

1. **Make your changes** in `internal/` or `cmd/`
2. **Write tests** in corresponding `*_test.go` files
3. **Run tests**:
   ```bash
   make test
   ```
4. **Run linter**:
   ```bash
   golangci-lint run
   ```

### Frontend Development

1. **Make your changes** in `frontend/src/`
2. **Run linter**:
   ```bash
   cd frontend && npm run lint
   ```
3. **Test in browser**: Development server at http://localhost:5173

### Database Migrations

When adding database changes:

1. Create a new migration file:
   ```bash
   # Format: YYYYMMDDHHmmss_description.go
   touch internal/database/migrations/20240115120000_add_feature.go
   ```

2. Implement Up and Down methods:
   ```go
   package migrations

   import "database/sql"

   type AddFeature struct{}

   func (m *AddFeature) Name() string {
       return "20240115120000_add_feature"
   }

   func (m *AddFeature) Up(tx *sql.Tx) error {
       // Migration logic
       return nil
   }

   func (m *AddFeature) Down(tx *sql.Tx) error {
       // Rollback logic
       return nil
   }
   ```

3. Register in `internal/database/migrations/registry.go`
4. Test migration:
   ```bash
   make build
   ./bin/server  # Runs migrations automatically
   ```

## Style Guides

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line
- Consider starting the commit message with an applicable emoji:
  - ✨ `:sparkles:` when adding a new feature
  - 🐛 `:bug:` when fixing a bug
  - 📝 `:memo:` when writing docs
  - 🎨 `:art:` when improving the format/structure of the code
  - ⚡ `:zap:` when improving performance
  - 🔒 `:lock:` when dealing with security
  - ✅ `:white_check_mark:` when adding tests

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Run `gofmt` before committing
- Use meaningful variable names
- Write godoc comments for exported functions
- Keep functions small and focused
- Handle errors explicitly

Example:

```go
// GetUserByID retrieves a user by their ID.
// Returns nil if user is not found.
func GetUserByID(id string) (*User, error) {
    user := &User{}
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(
        &user.ID, &user.Name, &user.Email,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

### TypeScript/React Style Guide

- Follow the ESLint configuration
- Use functional components with hooks
- Use TypeScript for type safety
- Keep components small and focused
- Use meaningful prop names

Example:

```typescript
interface LogEntryProps {
  log: Log;
  onSelect: (id: string) => void;
}

export function LogEntry({ log, onSelect }: LogEntryProps) {
  return (
    <div onClick={() => onSelect(log.id)}>
      <span className={levelColor[log.level]}>
        {log.level}
      </span>
      <p>{log.message}</p>
    </div>
  );
}
```

### Testing Guidelines

#### Backend Tests

```go
func TestUserRepository_Create(t *testing.T) {
    // Setup
    db := database.NewTestDB(t)
    database.RunTestMigrations(t, db, migrations.GetAll())
    defer db.Close()

    repo := models.NewUserRepository(db)

    // Execute
    user := &models.User{
        Username: "testuser",
        Password: "password123",
        Name:     "Test User",
        Role:     models.RoleUser,
    }
    err := repo.Create(user)

    // Assert
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    if user.ID == "" {
        t.Error("Expected user ID to be set")
    }
}
```

#### Frontend Tests

```typescript
import { render, screen } from '@testing-library/react';
import { LogEntry } from './LogEntry';

test('renders log entry', () => {
  const log = {
    id: '1',
    level: 'error',
    message: 'Test error',
    timestamp: new Date(),
  };

  render(<LogEntry log={log} onSelect={() => {}} />);

  expect(screen.getByText('error')).toBeInTheDocument();
  expect(screen.getByText('Test error')).toBeInTheDocument();
});
```

## Pull Request Process

1. **Update documentation** with details of changes to the interface, including new environment variables, exposed ports, useful file locations, and container parameters.

2. **Add tests** for any new functionality. The PR will not be merged until tests pass.

3. **Update the README.md** with details of changes if applicable.

4. **Ensure CI/CD passes**. All tests must pass and linting must be clean.

5. **Request review** from maintainers.

6. **Squash commits** before merging if requested.

### PR Checklist

- [ ] Code follows the style guidelines
- [ ] Self-review of code performed
- [ ] Code is commented where necessary
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added that prove fix/feature works
- [ ] All tests passing locally
- [ ] Dependent changes merged and published

## Release Process

Maintainers will handle releases using semantic versioning:

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

## Questions?

Don't hesitate to ask! You can:

- Open an issue with your question
- Join our Discord community
- Email the maintainers

## Recognition

Contributors will be recognized in:

- GitHub contributors list
- Release notes
- README acknowledgments

Thank you for contributing! 🎉
