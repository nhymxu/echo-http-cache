# echo-http-cache

This is a high performance Golang HTTP middleware for server-side application layer caching, ideal for REST APIs.

It is simple, superfast, thread safe and gives the possibility to choose the adapter (memory, Redis, DynamoDB etc).

The memory adapter minimizes GC overhead to near zero and supports some options of caching algorithms (LRU, MRU, LFU, MFU). This way, it is able to store plenty of gigabytes of responses, keeping great performance and being free of leaks.

Fork from [victorspringer/http-cache](https://github.com/victorspringer/http-cache) to support native Labstack Echo middleware and enhance some function

## Getting Started

### Installation
`go get github.com/nhymxu/echo-http-cache`

### Usage
This is an example of use with the memory adapter:

```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nhymxu/echo-http-cache"
	"github.com/nhymxu/echo-http-cache/adapter/memory"
)

func example(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}

func main() {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		log.Fatal(err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(10 * time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		log.Fatal(err)
	}

	handler := http.HandlerFunc(example)

	http.Handle("/", cacheClient.Middleware(handler))
	http.ListenAndServe(":8080", nil)
}
```

Example of Client initialization with Redis adapter:
```go
import (
    "github.com/nhymxu/echo-http-cache"
    "github.com/nhymxu/echo-http-cache/adapter/redis"
)

...

    ringOpt := &redis.RingOptions{
        Addrs: map[string]string{
            "server": ":6379",
        },
    }
    cacheClient, err := cache.NewClient(
    cache.ClientWithAdapter(redis.NewAdapter(ringOpt)),
    cache.ClientWithTTL(10 * time.Minute),
    cache.ClientWithRefreshKey("opn"),
)

...
```

## Benchmarks
The benchmarks were based on [allegro/bigache](https://github.com/allegro/bigcache) tests and used to compare it with the http-cache memory adapter.<br>
The tests were run using an Intel i5-2410M with 8GB RAM on Arch Linux 64bits.<br>
The results are shown below:

### Writes and Reads
```bash
cd adapter/memory/benchmark
go test -bench=. -benchtime=10s ./... -timeout 30m

BenchmarkHTTPCacheMamoryAdapterSet-4             5000000     343 ns/op    172 B/op    1 allocs/op
BenchmarkBigCacheSet-4                           3000000     507 ns/op    535 B/op    1 allocs/op
BenchmarkHTTPCacheMamoryAdapterGet-4            20000000     146 ns/op      0 B/op    0 allocs/op
BenchmarkBigCacheGet-4                           3000000     343 ns/op    120 B/op    3 allocs/op
BenchmarkHTTPCacheMamoryAdapterSetParallel-4    10000000     223 ns/op    172 B/op    1 allocs/op
BenchmarkBigCacheSetParallel-4                  10000000     291 ns/op    661 B/op    1 allocs/op
BenchmarkHTTPCacheMemoryAdapterGetParallel-4    50000000    56.1 ns/op      0 B/op    0 allocs/op
BenchmarkBigCacheGetParallel-4                  10000000     163 ns/op    120 B/op    3 allocs/op
```
http-cache writes are slightly faster and reads are much more faster.

### Garbage Collection Pause Time
```bash
cache=http-cache go run benchmark_gc_overhead.go

Number of entries:  20000000
GC pause for http-cache memory adapter:  2.445617ms

cache=bigcache go run benchmark_gc_overhead.go

Number of entries:  20000000
GC pause for bigcache:  7.43339ms
```
http-cache memory adapter takes way less GC pause time, that means smaller GC overhead.

## Roadmap
- Make it compliant with RFC7234
- Add more middleware configuration (cacheable status codes, paths etc)
- Develop gRPC middleware
- Develop Badger adapter
- Develop DynamoDB adapter
- Develop MongoDB adapter

## Godoc Reference
- [http-cache](https://godoc.org/github.com/nhymxu/echo-http-cache)
- [Memory adapter](https://godoc.org/github.com/nhymxu/echo-http-cache/adapter/memory)
- [Redis adapter](https://godoc.org/github.com/nhymxu/echo-http-cache/adapter/redis)

## License
http-cache is released under the [MIT License](https://github.com/github.com/nhymxu/echo-http-cache/blob/master/LICENSE).

Fork from [victorspringer/http-cache](https://github.com/victorspringer/http-cache) to support native Labstack Echo middleware and enhance some function
