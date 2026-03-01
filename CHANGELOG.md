# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Complete v1.0 Rewrite**: Introduced Generics (`Cache[T any]`) for 100% type safety.
- **Cache-Aside Helper**: Implemented `gocachex.Remember` with native cache stampede protection (`singleflight`).
- **New Directory Layout**: Backends are now top-level (`memory`, `redis`, `memcached`), minimizing nested abstractions.
- **Prometheus Metrics**: Included optional Prometheus metric wrappers for cache hit/miss/latency tracking.
- **Benchmarks**: Added comprehensive test benchmarks comparing the different backends.
- **CI/CD**: Added GitHub Actions pipeline for testing and coverage.

### Removed
- Removed old `interface{}` based `pkg/` structure.
- Removed custom overly-complex compression and sharding internal packages in favor of native driver support and functional boundaries.

### Changed
- Refactored `memory` backend to use a dedicated garbage collection routine for expired items.
- Redis backend now uses `github.com/redis/go-redis/v9`.
