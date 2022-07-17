package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/hinha/httpcache"
)

func newClient(client *http.Client, ICache httpcache.CacheInterface, webStaticFile bool) *httpcache.CacheHandler {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	newCacheHandler := httpcache.NewCacheHandlerRoundtrip(client.Transport, ICache, webStaticFile)
	client.Transport = newCacheHandler
	return newCacheHandler
}

// NewWithRedisCache will create a complete cache-support of HTTP client with using redis cache.
func NewRedisCache(client *http.Client, webStaticFile bool, options *httpcache.RedisCacheOptions, duration time.Duration) *httpcache.CacheHandler {
	c := redis.NewClient(&redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
		DB:       options.DB,
	})

	return newClient(client, NewCache(context.Background(), c, duration), webStaticFile)
}
