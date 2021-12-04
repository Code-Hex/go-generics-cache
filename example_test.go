package cache_test

import (
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

func ExampleCache() {
	// use simple cache algorithm without options.
	c := cache.New[string, int]()
	c.Set("a", 1)
	gota, aok := c.Get("a")
	gotb, bok := c.Get("b")
	fmt.Println(gota, aok)
	fmt.Println(gotb, bok)
	// Output:
	// 1 true
	// 0 false
}

func ExampleCacheWithExpiration() {
	c := cache.New(cache.AsFIFO[string, int]())
	exp := 250 * time.Millisecond
	c.Set("a", 1, cache.WithExpiration(exp))

	// check item is set.
	gota, aok := c.Get("a")
	fmt.Println(gota, aok)

	// set again
	c.Set("a", 2, cache.WithExpiration(exp))
	gota2, aok2 := c.Get("a")
	fmt.Println(gota2, aok2)

	// waiting expiration.
	time.Sleep(exp + 100*time.Millisecond) // + buffer

	gota3, aok3 := c.Get("a") // expired
	fmt.Println(gota3, aok3)
	// Output:
	// 1 true
	// 2 true
	// 0 false
}

func ExampleDelete() {
	c := cache.New(cache.AsLRU[string, int]())
	c.Set("a", 1)
	c.Delete("a")
	gota, aok := c.Get("a")
	fmt.Println(gota, aok)
	// Output:
	// 0 false
}

func ExampleKeys() {
	c := cache.New(cache.AsLFU[string, int]())
	c.Set("a", 1)
	c.Set("b", 1)
	c.Set("c", 1)
	fmt.Println(c.Keys())
	// Output:
	// [a b c]
}

func ExampleContains() {
	c := cache.New(cache.AsLRU[string, int]())
	c.Set("a", 1)
	fmt.Println(c.Contains("a"))
	fmt.Println(c.Contains("b"))
	// Output:
	// true
	// false
}

func ExampleNumberCache() {
	nc := cache.NewNumber[string, int]()
	nc.Set("a", 1)
	nc.Set("b", 2, cache.WithExpiration(time.Minute))
	av := nc.Increment("a", 1)
	gota, aok := nc.Get("a")

	bv := nc.Decrement("b", 1)
	gotb, bok := nc.Get("b")

	// not set keys
	cv := nc.Increment("c", 100)
	dv := nc.Decrement("d", 100)
	fmt.Println(av, gota, aok)
	fmt.Println(bv, gotb, bok)
	fmt.Println(cv)
	fmt.Println(dv)
	// Output:
	// 2 2 true
	// 1 1 true
	// 100
	// -100
}
