# Testing Documentation

## Test Suite Overview

The FreeStealer project includes comprehensive test coverage across all major components.

## Running Tests

```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test ./... -v

# Run tests with coverage report
go test ./... -cover

# Run tests for specific package
go test ./models -v
go test ./handlers -v
go test ./database -v
```

## Test Coverage

- **Database Package**: 87.5% coverage
- **Handlers Package**: 43.1% coverage
- **Models Package**: All tests passing

## Test Structure

### 1. Model Tests (`models/models_test.go`)

Tests for database models and their behavior:

- **User Model**
  - User creation
  - Unique username constraint
  - Unique email constraint
  - GitHub OAuth fields

- **Tier Model**
  - Tier creation
  - Default values (IsPublic, counters)
  - User relations

- **Vote Model**
  - Vote creation
  - Unique constraint (user_id + tier_id)
  - Valid vote types (upvote: 1, downvote: -1)

- **Comment Model**
  - Comment creation
  - Maximum length validation (100 characters)
  - Timestamp tracking
  - Relations with users and tiers

- **Soft Delete**
  - Soft delete functionality
  - DeletedAt field behavior

### 2. Handler Tests (`handlers/handlers_test.go`)

Tests for HTTP API endpoints:

- **User Handlers**
  - Create user (POST /users)
  - Get users (GET /users)
  - Input validation
  - HTTP method validation

- **Tier Handlers**
  - Create tier (POST /tiers)
  - Get tiers (GET /tiers)
  - Filter by platform
  - Filter by user_id
  - Public/private tier visibility

- **Vote Handlers**
  - Create upvote
  - Create downvote
  - Toggle vote (same vote removes it, different vote updates)
  - Invalid vote type validation

- **Comment Handlers**
  - Create comment
  - Get comments for tier
  - Length validation (max 100 chars)
  - Empty comment validation

### 3. Database Tests (`database/database_test.go`)

Tests for database operations and integrity:

- **Database Initialization**
  - Memory database creation
  - Table creation
  - Index creation

- **CRUD Operations**
  - Create and read operations
  - Foreign key relations
  - Cascade behavior (soft deletes)

- **Transactions**
  - Rollback functionality
  - Commit functionality

- **Query Optimization**
  - Index usage
  - Pagination
  - Ordering

- **Constraints**
  - Unique constraints (username, email)
  - Unique composite constraints (user_id + tier_id for votes)
  - Required field validation

## Test Data Setup

All tests use in-memory SQLite databases (`":memory:"`) for:
- Fast execution
- Isolation between tests
- No cleanup required
- Consistent test environment

## Key Testing Patterns

### 1. httptest for HTTP Handlers
```go
req := httptest.NewRequest(http.MethodGet, "/tiers", nil)
w := httptest.NewRecorder()
GetTiers(w, req)
```

### 2. In-Memory Database for Tests
```go
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
db.AutoMigrate(&models.User{}, &models.Tier{})
```

### 3. Subtests for Organization
```go
t.Run("Valid user creation", func(t *testing.T) {
    // Test code
})
```

## Known Limitations

1. **GitHubID Unique Index**: Uses partial index (only non-empty values) to allow multiple users without GitHub OAuth
2. **IsPublic Default**: Database default is `true`, requires explicit update for `false` values
3. **NOT NULL Constraints**: Validated at API layer rather than database layer
4. **Auth Package**: Not yet tested (0% coverage) - requires session/OAuth mocking

## Future Test Improvements

- [ ] Add auth package tests with mocked GitHub OAuth
- [ ] Increase handler test coverage to >70%
- [ ] Add integration tests for complete workflows
- [ ] Add performance/load tests
- [ ] Add API contract tests (OpenAPI validation)
- [ ] Add end-to-end tests with real database

## Test Best Practices

1. **Isolation**: Each test creates its own database
2. **Cleanup**: In-memory databases are automatically cleaned up
3. **Descriptive Names**: Test names clearly describe what they test
4. **Arrange-Act-Assert**: Tests follow AAA pattern
5. **Error Checking**: All database/HTTP operations check for errors
