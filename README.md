# Inject-able HTTP cache in Golang
 Just inject the cache to the HTTP client, this will integrate with cache (Redis)

## Quickstart

Make sure you have Go installed ([download](https://golang.org/dl/)). Version `1.16` or higher is required.

Initialize your project by creating a folder and then running `go mod init github.com/your/repo` ([learn more](https://blog.golang.org/using-go-modules)) inside the folder. Then install httpcache library with the [`go get`](https://golang.org/cmd/go/#hdr-Add_dependencies_to_current_module_and_install_them) command:

```sh
go get -u github.com/hinha/httpcache
```

## Example
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hinha/httpcache"
	"github.com/hinha/httpcache/cache"
)

var (
	url   = "https://google.com"
	token = "toke"
)

func main() {
	client := &http.Client{}
	// when webStatic is true can only be used static web ex: google.com
	// cause need cache-control
	webStatic := true
	_ = cache.NewRedisCache(client, &httpcache.RedisCacheOptions{
		Addr: "localhost:6379",
	}, time.Second*time.Duration(60))

	header := http.Header{}
	// header.Set("Authorization", token) // if need header

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal((err))
		}
		req.Header = header

		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Status Code", res.StatusCode)
	}
}
```

## Inspirations and Thanks
- [pquerna/cachecontrol](https://github.com/pquerna/cachecontrol) for the Cache-Header Extraction