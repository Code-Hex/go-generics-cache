package cache_test

import (
	"fmt"
	"sort"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

func ExampleCache() {
	c := cache.New[string, int]()
	c.Set("a", 1)
	c.Set("b", 2, cache.WithExpiration(time.Hour))
	av, aok := c.Get("a")
	bv, bok := c.Get("b")
	cv, cok := c.Get("c")
	fmt.Println(av, aok)
	fmt.Println(bv, bok)
	fmt.Println(cv, cok)
	// Output:
	// 1 true
	// 2 true
	// 0 false
}

func ExampleCacheKeys() {
	c := cache.New[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	keys := c.Keys()
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Println(key)
	}
	// Output:
	// a
	// b
	// c
}

func ExampleNumberCache() {
	c := cache.NewNumber[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
	av, aerr := c.Increment("a", 1)
	gota, aok := c.Get("a")

	bv, berr := c.Decrement("b", 1)
	gotb, bok := c.Get("b")

	// not set keys
	cv, cerr := c.Increment("c", 100)
	dv, derr := c.Decrement("d", 100)
	fmt.Println(av, aerr, gota, aok)
	fmt.Println(bv, berr, gotb, bok)
	fmt.Println(cv, cerr)
	fmt.Println(dv, derr)
	// Output:
	// 2 <nil> 2 true
	// 1 <nil> 1 true
	// 0 key[c]: not found item
	// 0 key[d]: not found item
}
