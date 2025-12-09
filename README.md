# Taskcore

The Open Source Alternative to Atlassian.

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Install golang-migrate

**Windows:**
```bash
# Using scoop
scoop install migrate

# Or download from: https://github.com/golang-migrate/migrate/releases
```

**Linux/macOS:**
```bash
# Using curl
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Or using brew
brew install golang-migrate
```

### Database Setup

1. Create a PostgreSQL database:
```sql
CREATE DATABASE taskcore;
```

2. Set your database connection string:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/taskcore?sslmode=disable"
```

3. Run migrations:
```bash
migrate -path migrations -database $DATABASE_URL up
```

### Migration Commands

**Apply all migrations:**
```bash
migrate -path migrations -database $DATABASE_URL up
```

**Rollback last migration:**
```bash
migrate -path migrations -database $DATABASE_URL down 1
```

**Check migration version:**
```bash
migrate -path migrations -database $DATABASE_URL version
```

**Force version (if stuck):**
```bash
migrate -path migrations -database $DATABASE_URL force VERSION
```

## Project Structure

```
taskcore/
├── cmd/taskcore/          # Main application entry point
├── internal/              # Private application code
│   ├── handler/          # HTTP handlers
│   ├── service/          # Business logic
│   ├── repository/       # Data access layer
│   ├── middleware/       # HTTP middleware
│   └── config/           # Configuration
├── migrations/           # SQL migrations (centralized)
├── web/                  # Frontend (React + Vite)
├── docs/                 # Documentation
└── scripts/              # Build and utility scripts
```

## Development

Coming soon...

## License

MIT