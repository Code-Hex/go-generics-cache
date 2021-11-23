# go-generics-cache

[![.github/workflows/test.yml](https://github.com/Code-Hex/go-generics-cache/actions/workflows/test.yml/badge.svg)](https://github.com/Code-Hex/go-generics-cache/actions/workflows/test.yml) [![codecov](https://codecov.io/gh/Code-Hex/go-generics-cache/branch/main/graph/badge.svg?token=Wm7UEwgiZu)](https://codecov.io/gh/Code-Hex/go-generics-cache)

go-generics-cache is an in-memory key:value store/cache that is suitable for applications running on a single machine. This in-memory cache uses [Go Generics](https://go.dev/blog/generics-proposal) which will be introduced in 1.18.

- a thread-safe
- implemented with [Go Generics](https://go.dev/blog/generics-proposal)
- Simple `map[string]interface{}` with expiration times
  - See [examples](https://github.com/Code-Hex/go-generics-cache/blob/main/simple/example_test.go)
- LRU cache
  - See [examples](https://github.com/Code-Hex/go-generics-cache/blob/main/lru/example_test.go)

## Requirements

Go 1.18 or later.

If Go 1.18 has not been released but you want to try this package, you can easily do so by using the [`gotip`](https://pkg.go.dev/golang.org/dl/gotip) command.

```sh
$ go install golang.org/dl/gotip@latest
$ gotip download # latest commit
$ gotip version
go version devel go1.18-c2397905e0 Sat Nov 13 03:33:55 2021 +0000 darwin/arm64
```

## Install

    $ go get github.com/Code-Hex/go-generics-cache

## Usage

See also [examples](https://github.com/Code-Hex/go-generics-cache/blob/main/example_test.go)

```go
package main

import (
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/simple"
)

func main() {
	// Create a simple cache. key as string, value as int.
	simpleCache := simple.New[string, int](simple.WithExpiration(time.Hour))

	// Create a cache for Number constraint. key as string, value as int.
	nc := cache.NewNumber[string, int](simpleCache)
	nc.Set("age", 26)

	// This will be compile error, because string is not satisfied cache.Number constraint.
	// nc := cache.NewNumber[string, string](simpleCache)

	incremented := nc.Increment("age", 1)
	fmt.Println(incremented) // 27

	decremented := nc.Decrement("age", 1)
	fmt.Println(decremented) // 26
}
```

## Articles

- English: [Some tips and bothers for Go 1.18 Generics](https://dev.to/codehex/some-tips-and-bothers-for-go-118-generics-lc7)
- Japanese: [Go 1.18 の Generics を使ったキャッシュライブラリを作った時に見つけた tips と微妙な点](https://zenn.dev/codehex/articles/3e6935ee6d853e)
