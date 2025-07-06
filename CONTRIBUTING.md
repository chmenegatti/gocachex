# Contributing to GoCacheX

First off, thank you for considering contributing to GoCacheX! It's people like you that make GoCacheX such a great tool.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Style Guidelines](#style-guidelines)
- [Testing](#testing)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## Getting Started

- Make sure you have a [GitHub account](https://github.com/signup/free)
- Fork the repository on GitHub
- Read the [README](README.md) to understand the project structure
- Look at the existing [issues](https://github.com/chmenegatti/gocachex/issues) to see if your contribution idea already exists

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to see if the problem has already been reported. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps which reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed after following the steps**
- **Explain which behavior you expected to see instead and why**
- **Include details about your configuration and environment**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate the steps**
- **Describe the current behavior and explain which behavior you expected to see instead**
- **Explain why this enhancement would be useful**

### Your First Code Contribution

Unsure where to begin contributing? You can start by looking through these issues:

- `good-first-issue` - issues which should only require a few lines of code
- `help-wanted` - issues which should be a bit more involved than beginner issues

## Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR-USERNAME/gocachex.git
   cd gocachex
   ```

2. **Install dependencies**
   ```bash
   go mod download
   make deps
   ```

3. **Install development tools**
   ```bash
   make install-tools
   ```

4. **Run tests to verify setup**
   ```bash
   make test
   ```

5. **Start development servers (optional)**
   ```bash
   # For Redis integration tests
   docker run -d -p 6379:6379 redis:alpine
   
   # For Memcached integration tests
   docker run -d -p 11211:11211 memcached:alpine
   ```

## Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make your changes**
   - Write clean, readable code
   - Add tests for new functionality
   - Update documentation as needed

3. **Run the test suite**
   ```bash
   make test
   make test-integration  # if applicable
   make lint
   make fmt-check
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add amazing feature"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```

6. **Create a Pull Request**
   - Fill in the PR template
   - Link any related issues
   - Ensure all checks pass

### Pull Request Guidelines

- Keep PRs focused and atomic
- Write clear commit messages
- Include tests for new features
- Update documentation
- Follow the existing code style
- Ensure backward compatibility when possible

## Style Guidelines

### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Write godoc comments for public APIs
- Keep functions small and focused
- Handle errors appropriately
- Use interfaces where appropriate

### Code Structure

```go
// Package comment explaining the package purpose
package packagename

import (
    // Standard library imports first
    "context"
    "fmt"
    
    // Third-party imports
    "github.com/external/package"
    
    // Local imports
    "github.com/chmenegatti/gocachex/pkg/config"
)

// Public constants and variables
const (
    DefaultTimeout = 30 * time.Second
)

// Public types with godoc comments
type MyType struct {
    // Exported fields with comments
    Field string `json:"field"`
}

// Public functions with godoc comments
func NewMyType() *MyType {
    return &MyType{}
}
```

### Commit Message Format

```
type(scope): short description

Longer description if needed

Fixes #123
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run integration tests
make test-integration

# Run benchmarks
make benchmark
```

### Writing Tests

- Write unit tests for all new functions
- Use table-driven tests when appropriate
- Include edge cases and error scenarios
- Mock external dependencies
- Write integration tests for complex features

Example test structure:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if result != tt.expected {
                t.Errorf("MyFunction() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Documentation

- Update README.md for new features
- Add godoc comments for public APIs
- Include examples in documentation
- Update CHANGELOG.md for notable changes

## Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create a PR with version bump
4. Tag the release after merge
5. GitHub Actions will handle the rest

## Getting Help

- Check existing [issues](https://github.com/chmenegatti/gocachex/issues)
- Join our discussions
- Ask questions in issues with the `question` label

## Recognition

Contributors will be recognized in:
- CHANGELOG.md for their contributions
- GitHub contributors page
- Release notes for significant contributions

Thank you for contributing to GoCacheX! 🚀
