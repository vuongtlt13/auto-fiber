# Code Coverage Guide

## Quick Start

### Basic Coverage Check

```bash
go test -cover ./...
```

### Detailed Coverage Report

```bash
./coverage.sh
```

### Clean Up Coverage Files

```bash
./coverage.sh clean
```

## Coverage Files

The following files are generated during coverage analysis and are automatically ignored by git:

- `coverage.out` - Raw coverage data
- `coverage.html` - HTML coverage report (open in browser for visual view)
- `*.cover` - Alternative coverage format
- `*.coverprofile` - Alternative coverage profile format

## Coverage Commands

### 1. Basic Coverage

```bash
go test -cover ./...
```

Shows overall coverage percentage for each package.

### 2. Function-Level Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Shows coverage percentage for each function.

### 3. HTML Report

```bash
go tool cover -html=coverage.out -o coverage.html
```

Generates an HTML file that can be opened in a browser for visual coverage analysis.

### 4. Coverage with Specific Package

```bash
go test -cover ./parser.go
go test -cover ./middleware.go
```

## Current Coverage Status

- **Total Coverage**: ~36.8%
- **Well Covered**: Core parsing, middleware, and map parsing functions
- **Needs Improvement**: Documentation generation, group operations, and some handler functions

## Areas for Improvement

1. **HTTP Methods**: Add tests for PUT, DELETE, PATCH, HEAD, OPTIONS
2. **Documentation**: Test OpenAPI generation functions
3. **Group Operations**: Test group routing functionality
4. **Edge Cases**: Test error conditions and validation edge cases

## Tips

- Run `./coverage.sh clean` before committing to avoid accidentally committing coverage files
- Use the HTML report to visually identify uncovered code sections
- Focus on testing public API functions first
- Consider adding integration tests for complex workflows
