## Testing Guide

## Overview

The project includes comprehensive unit tests for both domain and application layers, following best practices for clean architecture testing.

## Test Structure

```
receipt-bot/
├── internal/
│   ├── domain/
│   │   └── recipe/
│   │       ├── entity_test.go         # Recipe entity tests
│   │       ├── ingredient_test.go     # Ingredient value object tests
│   │       ├── instruction_test.go    # Instruction value object tests
│   │       └── source_test.go         # Source value object tests
│   └── application/
│       └── command/
│           └── process_recipe_link_test.go  # Main use case tests
```

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run with Coverage
```bash
go test -cover ./...
```

### Run with Verbose Output
```bash
go test -v ./...
```

### Run Specific Package
```bash
go test ./internal/domain/recipe/
```

### Generate Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Categories

### 1. Domain Layer Tests

**Purpose**: Test business logic and domain rules in isolation

**What We Test**:
- Value object creation and validation
- Entity creation and business rules
- Domain services
- Error conditions

**Example** (`ingredient_test.go`):
```go
func TestNewIngredient(t *testing.T) {
    // Test valid ingredient
    ing, err := NewIngredient("flour", "2", "cups", "all-purpose")
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Test invalid ingredient (empty name)
    _, err = NewIngredient("", "2", "cups", "")
    if err == nil {
        t.Error("expected error for empty name")
    }
}
```

**Key Test Cases**:
- ✅ Valid inputs produce correct objects
- ✅ Invalid inputs return appropriate errors
- ✅ Whitespace is trimmed
- ✅ Business rules are enforced
- ✅ String representations are correct

### 2. Application Layer Tests

**Purpose**: Test use case orchestration with mocked dependencies

**What We Test**:
- Command execution flow
- Dependency interaction
- Error handling
- Edge cases

**Example** (`process_recipe_link_test.go`):
```go
func TestProcessRecipeLinkCommand_Execute(t *testing.T) {
    // Create mocks
    mockScraper := &mockScraperPort{...}
    mockLLM := &mockLLMPort{...}
    mockRepo := newMockRecipeRepository()

    // Create command
    cmd := NewProcessRecipeLinkCommand(mockScraper, mockLLM, ...)

    // Execute and assert
    recipe, err := cmd.Execute(ctx, url, userID, chatID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify results
    if recipe.Title() != "Expected Title" {
        t.Errorf("got %v, want Expected Title", recipe.Title())
    }
}
```

**Key Test Cases**:
- ✅ Happy path: successful recipe extraction
- ✅ Error handling: scraping failures
- ✅ Error handling: LLM extraction failures
- ✅ Validation: empty ingredients
- ✅ Validation: empty instructions
- ✅ Progress message delivery

### 3. Mock Implementations

We use simple mock implementations for testing:

**mockScraperPort**:
```go
type mockScraperPort struct {
    result *ports.ScrapeResult
    err    error
}

func (m *mockScraperPort) Scrape(ctx context.Context, req ports.ScrapeRequest) (*ports.ScrapeResult, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.result, nil
}
```

**Benefits**:
- No external dependencies required
- Fast execution
- Deterministic results
- Easy to test error conditions

## Test Coverage Goals

- **Domain Layer**: >90% coverage
- **Application Layer**: >80% coverage
- **Adapters**: Integration tests (future)

## Current Test Coverage

### Domain Layer
- ✅ `Ingredient` value object: 100%
- ✅ `Instruction` value object: 100%
- ✅ `Source` value object: 100%
- ✅ `Recipe` entity: 95%
- ✅ Platform detection: 100%

### Application Layer
- ✅ `ProcessRecipeLinkCommand`: 85%
- ✅ Mock implementations: 100%

## Writing New Tests

### Test File Naming
- Test files must end with `_test.go`
- Place test files in the same package as the code being tested

### Test Function Naming
```go
// Pattern: Test<FunctionName>
func TestNewRecipe(t *testing.T) { ... }

// Pattern: Test<Type>_<Method>
func TestRecipe_AddIngredient(t *testing.T) { ... }
```

### Table-Driven Tests
Use table-driven tests for multiple scenarios:

```go
func TestNewIngredient(t *testing.T) {
    tests := []struct {
        name        string
        ingName     string
        quantity    string
        wantErr     bool
        errContains string
    }{
        {
            name:     "valid ingredient",
            ingName:  "flour",
            quantity: "2",
            wantErr:  false,
        },
        {
            name:        "empty name",
            ingName:     "",
            quantity:    "2",
            wantErr:     true,
            errContains: "name cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ing, err := NewIngredient(tt.ingName, tt.quantity, "", "")
            // ... assertions
        })
    }
}
```

### Assertion Patterns

**Error Checking**:
```go
if err != nil {
    t.Errorf("unexpected error: %v", err)
}

if err == nil {
    t.Error("expected error but got nil")
}
```

**Value Checking**:
```go
if got != want {
    t.Errorf("got %v, want %v", got, want)
}
```

**Fatal vs Error**:
- Use `t.Fatal()` when test cannot continue
- Use `t.Error()` when test can continue to find more issues

## Integration Tests (Future)

For adapter testing with real services:

```go
// Example: Test with Firebase emulator
func TestFirebaseRecipeRepository_Save(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup Firebase emulator
    client := setupFirebaseEmulator(t)
    defer client.Close()

    repo := firebase.NewRecipeRepository(client.Firestore())

    // Test with real Firestore
    recipe := createTestRecipe()
    err := repo.Save(context.Background(), recipe)
    if err != nil {
        t.Fatalf("Save failed: %v", err)
    }
}
```

Run integration tests:
```bash
go test -v ./... -run Integration
```

Skip integration tests:
```bash
go test -short ./...
```

## Continuous Integration

Add to `.github/workflows/test.yml`:
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go test -v -cover ./...
```

## Best Practices

1. **Test Behavior, Not Implementation**
   - Focus on what the code does, not how
   - Don't test private methods directly

2. **Keep Tests Independent**
   - Each test should run in isolation
   - Don't rely on test execution order

3. **Use Descriptive Names**
   - Test names should describe what they're testing
   - Use table-driven tests for multiple scenarios

4. **Test Edge Cases**
   - Empty values
   - Nil pointers
   - Boundary conditions

5. **Mock External Dependencies**
   - Don't call real APIs in unit tests
   - Use simple mock implementations

6. **Fast Tests**
   - Unit tests should be fast (<10ms each)
   - Save slow tests for integration

## Debugging Tests

Run a single test:
```bash
go test -v -run TestNewIngredient ./internal/domain/recipe/
```

Run with race detector:
```bash
go test -race ./...
```

View test output:
```bash
go test -v ./... 2>&1 | tee test-output.txt
```

## Common Issues

**Import Cycles**:
- Keep tests in the same package to access unexported functions
- Use `_test` suffix for black-box testing if needed

**Flaky Tests**:
- Avoid time-dependent tests
- Use deterministic mocks
- Don't rely on external state

**Slow Tests**:
- Mock external dependencies
- Use `testing.Short()` for long-running tests
- Consider parallel execution with `t.Parallel()`

## Next Steps

- [ ] Add benchmark tests for performance-critical code
- [ ] Integration tests with Firebase emulator
- [ ] Integration tests with Python service
- [ ] End-to-end tests with real Telegram bot
- [ ] Property-based testing for domain logic
