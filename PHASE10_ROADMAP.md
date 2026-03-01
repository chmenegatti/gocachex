# Phase 10 — Roadmap for v1.0

## PR Title
`docs: establish roadmap for community contributions post v1.0`

## Motivation
With the massive foundational v1.0 rewrite complete (Generics, Simplified Structure, Helpers, Benchmarks), `gocachex` is stable for enterprise use. However, open-source projects die without forward momentum. We must publish a roadmap to signal to the community where PRs are most desired.

## Proposed Roadmap

### 1. Distributed Invalidation (Pub/Sub)
- **Goal**: Allow multiple backend containers (e.g. running the local `memory` backend) to subscribe to a Redis Pub/Sub channel. If Server A modifies `user:1`, it broadcasts an invalidation event so Server B purges its local `memory` copy.
- **Priority**: High (Crucial for scaling L1/L2 patterns without dirty states).

### 2. Pluggable Serialization Interfaces
- **Goal**: Currently, the `redis` and `memcached` drivers hardcode `encoding/json`. We should expose a `Serializer` functional option allowing users to inject `MsgPack` or `Protobuf` for 10x serialization speed improvements.
- **Priority**: High.

### 3. OpenTelemetry (OTel) Tracing
- **Goal**: A new package `tracing/otel.go` acting similar to `metrics/prometheus.go` that injects distributed Zipkin/Jaeger span traces so users can see cache latency inside their overarching request cascades.
- **Priority**: Medium.

## Conclusion
The architectural changes implemented in Phases 1-9 are considered **essential before v1.0** because they break backwards compatibility (specifically the introduction of Generics and removing `pkg/`). The roadmap features listed above are non-breaking additive improvements that the community can rally behind on the journey to `v1.5`.
