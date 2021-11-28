package cache_test

import (
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/simple"
)

func ExampleNumberCache() {
	c := simple.NewCache[string, int]()
	nc := cache.NewNumber[string, int](c)
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
