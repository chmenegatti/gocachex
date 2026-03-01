package benchmarks_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/chmenegatti/gocachex/memcached"
	"github.com/chmenegatti/gocachex/memory"
	"github.com/chmenegatti/gocachex/redis"
	goredis "github.com/redis/go-redis/v9"
)

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := memory.New[string]()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = cache.Set(ctx, key, "value", time.Minute)
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := memory.New[string]()
	ctx := context.Background()
	cache.Set(ctx, "key", "value", time.Minute)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(ctx, "key")
	}
}

func BenchmarkRedisCache_Set(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatal(err)
	}
	defer mr.Close()

	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	cache := redis.New[string](client)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = cache.Set(ctx, key, "value", time.Minute)
	}
}

func BenchmarkRedisCache_Get(b *testing.B) {
	mr, err := miniredis.Run()
	if err != nil {
		b.Fatal(err)
	}
	defer mr.Close()

	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	cache := redis.New[string](client)
	ctx := context.Background()
	cache.Set(ctx, "key", "value", time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(ctx, "key")
	}
}

func BenchmarkMemcachedCache_Set(b *testing.B) {
	client := memcache.New("localhost:11211")
	if err := client.Ping(); err != nil {
		b.Skipf("skipping memcached benchmark: %v", err)
	}

	cache := memcached.New[string](client)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = cache.Set(ctx, key, "value", time.Minute)
	}
}

func BenchmarkMemcachedCache_Get(b *testing.B) {
	client := memcache.New("localhost:11211")
	if err := client.Ping(); err != nil {
		b.Skipf("skipping memcached benchmark: %v", err)
	}

	cache := memcached.New[string](client)
	ctx := context.Background()
	cache.Set(ctx, "key", "value", time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(ctx, "key")
	}
}
