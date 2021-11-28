package simple_test

import (
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/simple"
)

func ExampleCache() {
	c := simple.NewCache[string, int]()
	c.Set("a", 1, cache.WithExpiration(time.Hour))
	c.Set("b", 2)
	av, aok := c.Get("a")
	bv, bok := c.Get("b")
	cv, cok := c.Get("c")
	fmt.Println(av, aok)
	fmt.Println(bv, bok)
	fmt.Println(cv, cok)
	c.Delete("a")
	_, aok2 := c.Get("a")
	if !aok2 {
		fmt.Println("key 'a' has been deleted")
	}
	// Output:
	// 1 true
	// 2 true
	// 0 false
	// key 'a' has been deleted
}

func ExampleCacheKeys() {
	c := simple.NewCache[string, int]()
	c.Set("foo", 1)
	c.Set("bar", 2)
	c.Set("baz", 3)
	keys := c.Keys()
	for _, key := range keys {
		fmt.Println(key)
	}
	// Output:
	// foo
	// bar
	// baz
}
