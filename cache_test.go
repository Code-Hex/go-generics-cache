package cache

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestHasExpired(t *testing.T) {
	cases := []struct {
		name      string
		exp       time.Duration
		createdAt time.Time
		current   time.Time
		want      bool
	}{
		// expiration == createdAt + exp
		{
			name: "item expiration is zero",
			want: false,
		},
		{
			name:      "item expiration > current time",
			exp:       time.Hour * 24,
			createdAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			current:   time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name:      "item expiration < current time",
			exp:       time.Hour * 24,
			createdAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			current:   time.Date(2009, time.November, 12, 23, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "item expiration == current time",
			exp:       time.Second,
			createdAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			current:   time.Date(2009, time.November, 10, 23, 0, 1, 0, time.UTC),
			want:      false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backup := nowFunc
			nowFunc = func() time.Time { return tc.current }
			defer func() { nowFunc = backup }()

			it := Item[int]{
				Expiration: tc.exp,
				CreatedAt:  tc.createdAt,
			}
			if got := it.HasExpired(); tc.want != got {
				t.Fatalf("want %v, but got %v", tc.want, got)
			}
		})
	}
}

func TestGetItemExpired(t *testing.T) {
	c := New[struct{}, int]()
	c.SetItem(struct{}{}, Item[int]{
		Value:      1,
		Expiration: time.Hour * 24,
		CreatedAt:  time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	})

	backup := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2009, time.November, 12, 23, 0, 0, 0, time.UTC)
	}
	defer func() { nowFunc = backup }()

	v, err := c.GetItem(struct{}{})
	if !errors.Is(err, ErrExpired) {
		t.Errorf("want error %v but got %v", ErrExpired, err)
	}
	zeroValItem := Item[int]{}
	if zeroValItem != v {
		t.Errorf("want %v but got %v", zeroValItem, v)
	}

}

func TestMultiThreadIncr(t *testing.T) {
	nc := NewNumber[string, int]()
	nc.Set("counter", 0)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			_, err := nc.Increment("counter", 1)
			if err != nil {
				t.Logf("err: %v", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if counter, _ := nc.Get("counter"); counter != 100 {
		t.Errorf("want %v but got %v", 100, counter)
	}
}

func TestMultiThreadDecr(t *testing.T) {
	nc := NewNumber[string, int]()
	nc.Set("counter", 100)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			_, err := nc.Decrement("counter", 1)
			if err != nil {
				t.Logf("err: %v", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if counter, _ := nc.Get("counter"); counter != 0 {
		t.Errorf("want %v but got %v", 0, counter)
	}
}
