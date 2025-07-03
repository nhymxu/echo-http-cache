package redis

import (
	"context"
	"time"

	redisCache "github.com/go-redis/cache/v9"
	cache "github.com/nhymxu/echo-http-cache"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

// Adapter is the memory adapter data structure.
type Adapter struct {
	store *redisCache.Cache
}

// RingOptions exports go-redis RingOptions type.
type RingOptions redis.RingOptions

// Get implements the cache Adapter interface Get method.
func (a *Adapter) Get(key uint64) ([]byte, bool) {
	var c []byte
	if err := a.store.Get(context.TODO(), cache.KeyAsString(key), &c); err == nil {
		return c, true
	}

	return nil, false
}

// Set implements the cache Adapter interface Set method.
func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) {
	err := a.store.Set(&redisCache.Item{
		Key:   cache.KeyAsString(key),
		Value: response,
		TTL:   expiration.Sub(time.Now()),
	})
	if err != nil {
		return
	}
}

// Release implements the cache Adapter interface Release method.
func (a *Adapter) Release(key uint64) {
	err := a.store.Delete(context.TODO(), cache.KeyAsString(key))
	if err != nil {
		return
	}
}

// NewAdapter initializes Redis adapter.
func NewAdapter(opt *RingOptions) cache.Adapter {
	ringOpt := redis.RingOptions(*opt)

	store := redisCache.New(&redisCache.Options{
		Redis: redis.NewRing(&ringOpt),
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
		LocalCache: redisCache.NewTinyLFU(1000, time.Minute),
	})

	return &Adapter{store: store}
}
