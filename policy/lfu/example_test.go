package lfu_test

import (
	"fmt"

	"github.com/Code-Hex/go-generics-cache/policy/lfu"
)

func ExampleNewCache() {
	c := lfu.NewCache[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
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

func ExampleCache_Keys() {
	c := lfu.NewCache[string, int]()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	keys := c.Keys()
	for _, key := range keys {
		fmt.Println(key)
	}
	// Output:
	// a
	// b
	// c
}
