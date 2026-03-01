# gocachex

<p align="center">
  <img src="https://img.shields.io/github/v/release/chmenegatti/gocachex" alt="Latest Release">
  <a href="https://godoc.org/github.com/chmenegatti/gocachex"><img src="https://godoc.org/github.com/chmenegatti/gocachex?status.svg" alt="GoDoc"></a>
  <img src="https://github.com/chmenegatti/gocachex/actions/workflows/ci.yml/badge.svg" alt="Build Status">
  <img src="https://img.shields.io/badge/go-1.21+-blue.svg" alt="Go Version">
</p>

**A production-grade, generic caching library for Go.**

`gocachex` provides a unified, zero-allocation type-safe caching abstraction with interchangeable backends. Stop casting `interface{}` and start using modern Go features.

## Features

- **Generics throughout**: 100% type-safe `Cache[T any]` interface. No more internal type assertions.
- **Interchangeable Backends**: 
  - `memory`: High-performance thread-safe local cache with automatic background eviction.
  - `redis`: Distributed cache via `go-redis/v9` with JSON serialization.
  - `memcached`: Out-of-the-box Memcached support.
- **Cache-Aside Helper**: `gocachex.Remember` handles Cache-Miss data-loading with **stampede protection** (`singleflight`).
- **Observability**: Built-in, optional Prometheus Hooks for latency, hits, and misses.

## Installation

```bash
go get github.com/chmenegatti/gocachex
```

*Requires Go 1.21 or later.*

## Quick Start

### 1. In-Memory Cache

```go
package main

import (
	"context"
	"fmt"
	"time"
	"github.com/chmenegatti/gocachex/memory"
)

type User struct {
	ID   int
	Name string
}

func main() {
	// Initialize the memory cache with automatic background cleanup
	cache := memory.New[User](memory.WithCleanupInterval[User](time.Minute))
	ctx := context.Background()

	// Set
	_ = cache.Set(ctx, "user:1", User{ID: 1, Name: "Alice"}, 5*time.Minute)

	// Get
	user, err := cache.Get(ctx, "user:1")
	if err == nil {
		fmt.Printf("Cached user: %+v\n", user)
	}
}
```

### 2. Cache-Aside Pattern with Stampede Protection

The `Remember` helper avoids executing your database query multiple times when a flurry of requests causes a cache miss.

```go
user, err := gocachex.Remember(ctx, cache, "user:1", 5*time.Minute, func() (User, error) {
    // ⬇️ This is only executed ONCE, even if 100 goroutines hit this simultaneously
    return db.GetUserByID(1)
})
```

### 3. Redis Cache

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chmenegatti/gocachex/redis"
	goredis "github.com/redis/go-redis/v9"
)

type Product struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func main() {
	rdb := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	
	// Initialize the generic Redis cache for 'Product'
	cache := redis.New[Product](rdb)
	defer cache.Close()

	ctx := context.Background()
	_ = cache.Set(ctx, "product:101", Product{ID: 101, Title: "Mechanical Keyboard"}, 10*time.Minute)

	product, _ := cache.Get(ctx, "product:101")
	fmt.Printf("Cached product: %+v\n", product)
}
```

### 4. Memcached Cache

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/chmenegatti/gocachex/memcached"
)

type Config struct {
	FeatureToggle bool `json:"feature_toggle"`
}

func main() {
	mc := memcache.New("localhost:11211")

	// Initialize the generic Memcached cache for 'Config'
	cache := memcached.New[Config](mc)
	ctx := context.Background()

	_ = cache.Set(ctx, "app:config", Config{FeatureToggle: true}, time.Hour)

	cfg, _ := cache.Get(ctx, "app:config")
	fmt.Printf("Cached config: %+v\n", cfg)
}
```

## Architecture Overview

`gocachex` revolves around a single core interface:

```go
type Cache[T any] interface {
	Get(ctx context.Context, key string) (T, error)
	Set(ctx context.Context, key string, value T, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Close() error
}
```

Backends live in their own packages (`memory`, `redis`, `memcached`) preventing heavy dependencies if you only need a specific backend.

## Performance Benchmarks

Run the benchmarks locally with `go test -bench . ./benchmarks/...`:

| Backend | Operation | Time (ns/op) |
| --- | --- | --- |
| **Memory** | Set | ~476 ns/op |
| **Memory** | Get | ~41 ns/op |
| **Redis** (Local) | Set | ~35,904 ns/op |
| **Redis** (Local) | Get | ~31,403 ns/op |

*Note: Distributed backends incorporate network and serialization latency.*

## Supported Backends
* `github.com/chmenegatti/gocachex/memory` - Native sync.RWMutex
* `github.com/chmenegatti/gocachex/redis` - Uses `github.com/redis/go-redis/v9` 
* `github.com/chmenegatti/gocachex/memcached` - Uses `github.com/bradfitz/gomemcache`

## Roadmap

Upcoming features planned for v2:
- [ ] Distributed cache invalidation (Pub/Sub)
- [ ] Pluggable serialization interfaces (MsgPack, Protobuf)
- [ ] OpenTelemetry integration hooks

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md).

## License

MIT License. See [LICENSE](LICENSE) for more information.
