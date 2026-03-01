# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Standardized documentation: `CONTRIBUTING.md`, `CHANGELOG.md`.
- GitHub issue and PR templates.
- Standalone top-level backend integration paths (`memory/`, `redis/`, `memcached/`).

### Changed
- Moved `pkg/` internals to the `internal/` directory to satisfy idiomatic Go encapsulation.
