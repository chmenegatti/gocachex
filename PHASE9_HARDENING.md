# Phase 9 — Production Hardening

## PR Title
`refactor: harden concurrency protocols and standardize domain errors`

## Motivation
Enterprise edge caches face unique threats: simultaneous thousand-goroutine reads (stampedes), backend network partitions, and massive serialization payloads. The library must be resilient to these extremes.

## Technical Improvements
1. **Concurrency Safety**: The `memory` backend is heavily guarded using `sync.RWMutex`. Specifically, TTL evaluation occurs inside the `RLock()`, categorically preventing dirty reads or panics (`fatal error: concurrent map read and map write`).
2. **Stampede Deduplication**: The `Remember` helper implements `golang.org/x/sync/singleflight`, guaranteeing that an expired key does not spawn duplicated loader executions, protecting the downstream database.
3. **Error Semantics**: Standardized `gocachex.ErrCacheMiss`. By translating driver errors (e.g., `go-redis.Nil` or `memcache.ErrCacheMiss`) to the domain error `ErrCacheMiss`, downstream user-applications write only *one* error handler regardless of what backend sits beneath the generic `gocachex`.

## Impact on the Project
- **Robustness**: Prevents cascade failures and database exhaustion.
- **Ecosystem Best Practices**: Zero leaked abstraction details; clients handle only standard `gocachex` errors.
