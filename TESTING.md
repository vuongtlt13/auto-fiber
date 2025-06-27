# Testing Guide for AutoFiber

This document provides a comprehensive guide for testing AutoFiber applications.

## Overview

AutoFiber comes with a comprehensive test suite that covers:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end API testing
- **Benchmark Tests**: Performance testing
- **Test Helpers**: Utility functions for testing

## Test Structure

```
├── app_test.go           # App creation and configuration tests
├── types_test.go         # Type and validation tests
├── parser_test.go        # Request parsing tests
├── validator_test.go     # Validation tests
├── routes_test.go        # Route registration tests
├── docs_test.go          # Documentation generation tests
├── integration_test.go   # End-to-end integration tests
├── benchmark_test.go     # Performance benchmarks
└── test_helpers.go       # Test utility functions
```

## Running Tests

### Run All Tests

```bash
go test ./...
```

### Run Specific Test Files

```bash
go test -v ./app_test.go
go test -v ./integration_test.go
```

### Run Tests with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Benchmark Tests

```bash
go test -bench=. ./benchmark_test.go
go test -bench=BenchmarkParseRequest -benchmem ./benchmark_test.go
```

## Test Categories

### 1. Unit Tests

#### App Tests (`app_test.go`)

Tests for application creation and configuration:

- App creation with default options
- App creation with custom configuration
- App startup and shutdown

#### Type Tests (`types_test.go`)

Tests for type definitions and validation:

- ParseSource enum functionality
- RouteConfig validation
- DocsConfig validation

#### Parser Tests (`parser_test.go`)

Tests for request parsing functionality:

- Body parsing (JSON)
- Query parameter parsing
- Path parameter parsing
- Header parsing
- Cookie parsing
- Form data parsing
- Multi-source parsing
- Error handling for invalid data

#### Validator Tests (`validator_test.go`)

Tests for validation functionality:

- Request validation
- Response validation
- Various validation rules (required, min, max, email, etc.)
- Nested struct validation

#### Route Tests (`routes_test.go`)

Tests for route registration:

- Single route registration
- Multiple routes registration
- Route groups
- Middleware integration
- Invalid route handling

#### Documentation Tests (`docs_test.go`)

Tests for OpenAPI documentation generation:

- Basic OpenAPI spec generation
- Schema generation from structs
- Validation rule mapping
- Multiple routes documentation

### 2. Integration Tests (`integration_test.go`)

#### Basic API Testing

```go
func TestIntegration_BasicAPI(t *testing.T) {
    // Tests complete API flow with request/response parsing and validation
}
```

#### Multi-Source Parsing Testing

```go
func TestIntegration_MultiSourceParsing(t *testing.T) {
    // Tests parsing from multiple sources (path, query, header, body)
}
```

#### Route Groups Testing

```go
func TestIntegration_RouteGroups(t *testing.T) {
    // Tests route group functionality
}
```

#### Documentation Testing

```go
func TestIntegration_Documentation(t *testing.T) {
    // Tests OpenAPI spec and Swagger UI endpoints
}
```

### 3. Benchmark Tests (`benchmark_test.go`)

Performance benchmarks for:

- Request parsing (body only)
- Request parsing (multi-source)
- Request validation
- Response validation
- Route registration
- HTTP request handling
- OpenAPI spec generation

## Test Helpers

### Creating Test Apps

```go
app, err := CreateTestApp(WithDocsConfig(DocsConfig{
    Title:   "Test API",
    Version: "1.0.0",
}))
```

### Making Test Requests

```go
req := TestRequest{
    Method: "POST",
    Path:   "/users",
    Body:   userData,
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
}

resp, err := MakeTestRequest(app, req)
```

### Creating Test Routes

```go
route := CreateTestRoute("/test", "GET", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "test"})
})

routeWithStructs := CreateTestRouteWithStructs(
    "/users",
    "POST",
    handler,
    &UserRequest{},
    &UserResponse{},
)
```

## Writing Custom Tests

### Example: Testing a Custom Endpoint

```go
func TestCustomEndpoint(t *testing.T) {
    // Create test app
    app, err := CreateTestApp()
    assert.NoError(t, err)

    // Define request/response structs
    type CreateOrderRequest struct {
        ProductID int     `json:"product_id" validate:"required"`
        Quantity  int     `json:"quantity" validate:"required,min=1"`
        Price     float64 `json:"price" validate:"required,min=0"`
    }

    type CreateOrderResponse struct {
        OrderID   int     `json:"order_id" validate:"required"`
        ProductID int     `json:"product_id" validate:"required"`
        Quantity  int     `json:"quantity" validate:"required"`
        Total     float64 `json:"total" validate:"required"`
    }

    // Create route
    route := RouteConfig{
        Path:   "/orders",
        Method: "POST",
        Handler: func(c *fiber.Ctx) error {
            var req CreateOrderRequest
            if err := ParseRequest(c, &req); err != nil {
                return c.Status(400).JSON(fiber.Map{"error": err.Error()})
            }

            if err := ValidateRequest(&req); err != nil {
                return c.Status(400).JSON(fiber.Map{"error": err.Error()})
            }

            response := CreateOrderResponse{
                OrderID:   1,
                ProductID: req.ProductID,
                Quantity:  req.Quantity,
                Total:     req.Price * float64(req.Quantity),
            }

            return c.JSON(response)
        },
        RequestStruct:  &CreateOrderRequest{},
        ResponseStruct: &CreateOrderResponse{},
    }

    // Register route
    err = RegisterRoute(app, route)
    assert.NoError(t, err)

    // Test valid request
    t.Run("valid request", func(t *testing.T) {
        requestBody := CreateOrderRequest{
            ProductID: 123,
            Quantity:  2,
            Price:     29.99,
        }

        req := TestRequest{
            Method: "POST",
            Path:   "/orders",
            Body:   requestBody,
            Headers: map[string]string{
                "Content-Type": "application/json",
            },
        }

        resp, err := MakeTestRequest(app, req)
        assert.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)

        var response CreateOrderResponse
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, requestBody.ProductID, response.ProductID)
        assert.Equal(t, requestBody.Quantity, response.Quantity)
        assert.Equal(t, requestBody.Price*float64(requestBody.Quantity), response.Total)
    })

    // Test invalid request
    t.Run("invalid request", func(t *testing.T) {
        requestBody := CreateOrderRequest{
            ProductID: 123,
            Quantity:  0, // Invalid: below minimum
            Price:     -10, // Invalid: negative price
        }

        req := TestRequest{
            Method: "POST",
            Path:   "/orders",
            Body:   requestBody,
            Headers: map[string]string{
                "Content-Type": "application/json",
            },
        }

        resp, err := MakeTestRequest(app, req)
        assert.NoError(t, err)
        assert.Equal(t, 400, resp.StatusCode)
    })
}
```

## Best Practices

### 1. Test Organization

- Group related tests using `t.Run()`
- Use descriptive test names
- Test both success and failure cases

### 2. Test Data

- Use realistic test data
- Test edge cases and boundary conditions
- Use table-driven tests for multiple scenarios

### 3. Assertions

- Use `testify/assert` for clear assertions
- Check both positive and negative cases
- Validate response structure and content

### 4. Performance

- Use benchmarks for performance-critical code
- Run benchmarks with `-benchmem` flag
- Monitor memory allocations

### 5. Coverage

- Aim for high test coverage
- Focus on critical paths
- Test error handling scenarios

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.23
      - run: go test -v -cover ./...
      - run: go test -bench=. ./benchmark_test.go
```

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure all dependencies are properly installed

   ```bash
   go mod tidy
   ```

2. **Test Failures**: Check if the underlying code has changed

   ```bash
   go test -v ./...
   ```

3. **Benchmark Failures**: Ensure consistent environment
   ```bash
   go test -bench=. -benchmem ./benchmark_test.go
   ```

### Debugging Tests

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test with debug info
go test -v -run TestSpecificFunction ./...

# Run tests with race detection
go test -race ./...
```

## Contributing

When adding new features to AutoFiber:

1. Write unit tests for new functionality
2. Add integration tests for end-to-end scenarios
3. Include benchmark tests for performance-critical code
4. Update this documentation if needed
5. Ensure all tests pass before submitting PR

## Test Coverage Goals

- **Unit Tests**: >90% coverage
- **Integration Tests**: All major features covered
- **Benchmark Tests**: All performance-critical functions
- **Error Handling**: All error paths tested
