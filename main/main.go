package main

import (
	"context"
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

type User struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Salary int    `json:"salary"`
}

type CacheObj[V interface{}] struct {
	Key   string
	Value V
}

func main() {
	// EncodeSimple()
	// LoadSimple()

	EncodeMRU()
	LoadMRU()
}

func LoadMRU() {
	ctx, cancel := context.WithCancel(context.Background())
	studentCache := cache.NewContext(ctx, cache.AsMRU[string, User](), cache.WithJanitorInterval[string, User](10*time.Minute))
	defer cancel()

	err := studentCache.Load("user_mru.gob")
	if err != nil {
		panic(err)
	}

	fmt.Println("-------- MRU ------")

	v, ok := studentCache.Get("u1")
	if ok {
		fmt.Printf("FOUND: %v", v)
	} else {
		fmt.Printf("NOT FOUND !!")
	}
	v, ok = studentCache.Get("u2")

	if ok {
		fmt.Printf("FOUND: %v", v)
	} else {
		fmt.Printf("NOT FOUND !!")
	}
	fmt.Println("------------")
}

func EncodeMRU() {
	s1 := User{Name: "ali", Age: 22, Salary: 200000}
	s2 := User{Name: "ashraf", Age: 22, Salary: 100000}

	ctx, cancel := context.WithCancel(context.Background())
	studentCache := cache.NewContext(ctx, cache.AsMRU[string, User](), cache.WithJanitorInterval[string, User](10*time.Minute))
	defer cancel()

	studentCache.Set("u1", s1, cache.WithExpiration(time.Hour))
	studentCache.Set("u2", s2, cache.WithExpiration(time.Hour))
	studentCache.Save("user_mru.gob")

}

func LoadSimple() {
	ctx, cancel := context.WithCancel(context.Background())
	studentCache := cache.NewContext(ctx, cache.WithJanitorInterval[string, User](10*time.Minute))
	defer cancel()

	err := studentCache.Load("user_simple.gob")
	if err != nil {
		panic(err)
	}
	fmt.Println("-------- Simple ------")
	v, ok := studentCache.Get("u1")
	if ok {
		fmt.Printf("FOUND: %v", v)
	} else {
		fmt.Printf("NOT FOUND !!")
	}
	v, ok = studentCache.Get("u2")

	if ok {
		fmt.Printf("FOUND: %v", v)
	} else {
		fmt.Printf("NOT FOUND !!")
	}

	fmt.Println("------------")
}

func EncodeSimple() {
	s1 := User{Name: "ali", Age: 22, Salary: 200000}
	s2 := User{Name: "ashraf", Age: 22, Salary: 100000}

	ctx, cancel := context.WithCancel(context.Background())
	studentCache := cache.NewContext(ctx, cache.WithJanitorInterval[string, User](10*time.Minute))
	defer cancel()

	studentCache.Set("u1", s1, cache.WithExpiration(time.Hour))
	studentCache.Set("u2", s2, cache.WithExpiration(time.Hour))
	studentCache.Save("user_simple.gob")

}
