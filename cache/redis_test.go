package cache

import (
	"context"
	"testing"
	"time"

	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"

	"github.com/hinha/httpcache"
)

func TestCacheRedis(t *testing.T) {
	testKey := "KEY"
	testVal := httpcache.CachedResponse{
		Response:      nil,
		RequestURI:    "http://platform.com",
		RequestMethod: "GET",
		CachedTime:    time.Now(),
	}

	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // default set
		DB:       0,  // default set
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock := redismock.NewNiceMock(c)
	cacheObj := NewCache(ctx, mock, 15*time.Second)

	// try to set
	mock.On("Set", ctx, testKey, string(testVal.ToByte()), 15*time.Second).Return(redis.NewStatusResult("set", nil))
	if err := cacheObj.Set(testKey, testVal); err != nil {
		t.Fatalf("expected %v, got %v", nil, err)
	}

	// try to get cache
	mock.On("Get", ctx, testKey).Return(redis.NewStringResult(string(testVal.ToByte()), nil))
	res, err := cacheObj.Get(testKey)
	if err != nil {
		t.Fatalf("expected %v, got %v", nil, err)
	}

	// assert the content
	if res.RequestURI != testVal.RequestURI {
		t.Fatalf("expected %v, got %v", testVal.RequestURI, res.RequestURI)
	}
	// assert the content
	if res.RequestMethod != testVal.RequestMethod {
		t.Fatalf("expected %v, got %v", testVal.RequestMethod, res.RequestMethod)
	}

	mock.On("Del", ctx, []string{testKey}).Return(redis.NewIntResult(1, nil))
	if err := cacheObj.Delete(testKey); err != nil {
		t.Fatalf("expected %v, got %v", nil, err)
	}
}
