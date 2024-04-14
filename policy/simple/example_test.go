package simple_test

import (
	"fmt"
	"testing"

	"github.com/Code-Hex/go-generics-cache/policy/simple"
)

func ExampleNewCache() {
	c := simple.NewCache[string, int]()
	c.Set("a", 1)
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
	// update
	c.Set("b", 3)
	newbv, _ := c.Get("b")
	fmt.Println(newbv)
	// Output:
	// 1 true
	// 2 true
	// 0 false
	// key 'a' has been deleted
	// 3
}

func ExampleCache_Keys() {
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

func ExampleCache_Len() {
	c := simple.NewCache[string, int]()
	c.Set("foo", 1)
	c.Set("bar", 2)
	c.Set("baz", 3)
	len := c.Len()
	fmt.Println(len)
	// Output:
	// 3
}

func BenchmarkLenWithKeys(b *testing.B) {
	c := simple.NewCache[string, int]()
	c.Set("foo", 1)
	c.Set("bar", 2)
	c.Set("baz", 3)
	for i := 0; i < b.N; i++ {
		var _ = len(c.Keys())
	}
}

func BenchmarkJustLen(b *testing.B) {
	c := simple.NewCache[string, int]()
	c.Set("foo", 1)
	c.Set("bar", 2)
	c.Set("baz", 3)
	for i := 0; i < b.N; i++ {
		var _ = c.Len()
	}
}
