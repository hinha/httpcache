package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/hinha/httpcache"
)

type redisCache struct {
	ctx        context.Context
	cache      redis.Cmdable
	expiryTime time.Duration
}

func NewCache(ctx context.Context, c redis.Cmdable, exptime time.Duration) httpcache.CacheInterface {
	return &redisCache{ctx: ctx, cache: c, expiryTime: exptime}
}

func (c *redisCache) Set(key string, value httpcache.CachedResponse) error {
	jsonByte, _ := json.Marshal(value)
	stat := c.cache.Set(c.ctx, key, string(jsonByte), c.expiryTime)
	if err := stat.Err(); err != nil {
		return httpcache.ErrStorageInternal
	}

	return nil
}

func (c *redisCache) Get(key string) (res httpcache.CachedResponse, err error) {
	stat := c.cache.Get(c.ctx, key)
	if err := stat.Err(); err == redis.Nil {
		// not found
		return res, httpcache.ErrStorageNotFound
	} else if err != nil {
		return res, httpcache.ErrStorageInternal
	}

	if err := json.Unmarshal([]byte(stat.Val()), &res); err != nil {
		return res, httpcache.ErrStorageInternal
	}

	return res, nil
}

func (c *redisCache) Delete(key string) error {
	stat := c.cache.Del(c.ctx, key)
	if err := stat.Err(); err != nil {
		return httpcache.ErrStorageInternal
	}
	return nil
}

func (c *redisCache) Flush() error {
	stat := c.cache.FlushAll(c.ctx)
	if err := stat.Err(); err != nil {
		return httpcache.ErrStorageInternal
	}

	return nil
}
