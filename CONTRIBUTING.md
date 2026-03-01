# Contributing to gocachex

Thank you for your interest in contributing to `gocachex`! This document provides guidelines and workflows for contributing to the library.

## Getting Started

1.  **Fork** the repository.
2.  **Clone** your fork locally.
3.  **Create a branch** for your feature or bug fix.

## Development Workflow

1.  Write idiomatic Go code.
2.  Write unit tests for your changes.
3.  Ensure your code passes the linting (`go vet`, `golangci-lint`) and testing (`go test ./...`).

## Adding a New Backend

Open an Issue first to discuss the implementation details. Ensure your backend fully implements the `gocachex.Cache` interface and keep serialization generic where appropriate.

## Pull Requests

Reference the Issue your PR solves in the description. Ensure GitHub Actions CI passes.
