package cache_test

import (
	"fmt"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/simple"
)

func ExampleNumberCache() {
	c := cache.NewNumber[string, int](simple.New[string, int]())
	c.Set("a", 1)
	c.Set("b", 2)
	av := c.Increment("a", 1)
	gota, aok := c.Get("a")

	bv := c.Decrement("b", 1)
	gotb, bok := c.Get("b")

	// not set keys
	cv := c.Increment("c", 100)
	dv := c.Decrement("d", 100)
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
