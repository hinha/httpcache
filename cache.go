package httpcache

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	// ErrStorageNotFound will throw when some not found key error in storage
	ErrStorageNotFound = errors.New("key is missing")
	// ErrStorageInternal will throw when some internal error in storage occurred
	ErrStorageInternal = errors.New("internal error in storage")
)

// RedisCacheOptions for storing data for Redis connections
type RedisCacheOptions struct {
	Addr     string
	Password string
	DB       int
}

// CacheInterface implement method cache
type CacheInterface interface {
	Set(key string, value CachedResponse) error
	Get(key string) (res CachedResponse, err error)
	Delete(key string) error
	Flush() error
}

// CachedResponse represent the cacher struct item
type CachedResponse struct {
	Response      []byte    `json:"response"`      // The dumped response body
	RequestURI    string    `json:"requestUri"`    // The requestURI of the response
	RequestMethod string    `json:"requestMethod"` // The HTTP Method that call the request for this response
	CachedTime    time.Time `json:"cachedTime"`    // The timestamp when this response is Cached
}

func (c CachedResponse) ToByte() []byte {
	data, _ := json.Marshal(c)
	return data
}
