# Contributing to gocachex

Thank you for your interest in contributing to `gocachex`! This document provides guidelines and workflows for contributing to the library.

## Getting Started

1.  **Fork** the repository.
2.  **Clone** your fork locally: `git clone https://github.com/your-username/gocachex.git`
3.  **Create a branch** for your feature or bug fix: `git checkout -b feature/my-new-feature` or `bugfix/issue-123`

## Development Workflow

1.  **Ensuring Code Quality**:
    *   Write idiomatic Go code.
    *   Write table-driven tests for your changes.
    *   Target +80% test coverage.

2.  **Running Tests**:
    ```bash
    go test -v -race -cover ./...
    ```

    *If you are testing Redis or Memcached locally, you must have them running on localhost at default ports or mock them.*

3.  **Running Benchmarks**:
    ```bash
    go test -bench . ./benchmarks/...
    ```

## Adding a New Backend

If you're proposing a new backend (e.g., BadgerDB, DynamoDB):
- Open an Issue first to discuss the implementation details.
- Ensure your backend fully implements the `gocachex.Cache[T]` generic interface.
- Keep serialization generic where appropriate.

## Pull Requests

1.  Reference the Issue your PR solves in the PR description (e.g. `Fixes #12`).
2.  Ensure GitHub Actions CI passes (Go Vet, Lint formatting, Tests).
3.  Fill out the Pull Request template comprehensively.

Thank you for contributing!
