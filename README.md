# go-generics-cache

go-generics-cache is an in-memory key:value store/cache that is suitable for applications running on a single machine. This in-memory cache uses [Go Generics](https://go.dev/blog/generics-proposal) which will be introduced in 1.18.

- implemented with [Go Generics](https://go.dev/blog/generics-proposal)
- a thread-safe `map[string]interface{}` with expiration times

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

playground: https://gotipplay.golang.org/p/FXRk6ngYV-s

```go
package main

import (
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

func main() {
	// Create a cache. key as string, value as int.
	c1 := cache.New[string, int]()

	// Sets the value of int. you can set with expiration option.
	c1.Set("foo", 1, cache.WithExpiration(time.Hour))

	// the value never expires.
	c1.Set("bar", 2)

	foo, ok := c1.Get("foo")
	if ok {
		fmt.Println(foo) // 1
	}

	fmt.Println(c1.Keys()) // outputs "foo" "bar" may random

	// Create a cache. key as int, value as string.
	c2 := cache.New[int, string]()
	c2.Set(1, "baz")
	baz, ok := c2.Get(1)
	if ok {
		fmt.Println(baz) // "baz"
	}

	// Create a cache for Number constraint.. key as string, value as int.
	nc := cache.NewNumber[string, int]()
	nc.Set("age", 26)

	// This will be compile error, because string is not satisfied cache.Number constraint.
	// nc := cache.NewNumber[string, string]()

	incremented, _ := nc.Increment("age", 1)
	fmt.Println(incremented) // 27

	decremented, _ := nc.Decrement("age", 1)
	fmt.Println(decremented) // 26
}
```