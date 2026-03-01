package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/chmenegatti/gocachex"
	"github.com/chmenegatti/gocachex/redis"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisCache[T any](t *testing.T) (*redis.RedisCache[T], *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cache := redis.New[T](client)
	return cache, mr
}

type customStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRedisCache_BasicOperations(t *testing.T) {
	ctx := context.Background()

	cache, mr := setupRedisCache[customStruct](t)
	defer mr.Close()
	defer cache.Close()

	data := customStruct{Name: "test", Value: 42}

	// Set
	err := cache.Set(ctx, "k1", data, time.Minute)
	require.NoError(t, err)

	// Get
	val, err := cache.Get(ctx, "k1")
	require.NoError(t, err)
	assert.Equal(t, data, val)

	// Exists
	exists, err := cache.Exists(ctx, "k1")
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete
	err = cache.Delete(ctx, "k1")
	require.NoError(t, err)

	_, err = cache.Get(ctx, "k1")
	assert.ErrorIs(t, err, gocachex.ErrCacheMiss)
}

func TestRedisCache_Clear(t *testing.T) {
	ctx := context.Background()

	cache, mr := setupRedisCache[string](t)
	defer mr.Close()

	cache.Set(ctx, "a", "1", 0)
	cache.Set(ctx, "b", "2", 0)

	err := cache.Clear(ctx)
	require.NoError(t, err)

	exists, _ := cache.Exists(ctx, "a")
	assert.False(t, exists)
}
