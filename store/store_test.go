package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestSetGet(t *testing.T) {
	s := New()
	s.Set("foo", "bar")

	v, ok := s.Get("foo")
	if !ok || v != "bar" {
		t.Fatalf("Get(foo) = %q, %v; want \"bar\", true", v, ok)
	}
}

func TestGetMissing(t *testing.T) {
	s := New()

	v, ok := s.Get("missing")
	if ok || v != "" {
		t.Fatalf("Get(missing) = %q, %v; want \"\", false", v, ok)
	}
}

func TestOverwrite(t *testing.T) {
	s := New()
	s.Set("foo", "one")
	s.Set("foo", "two")

	v, ok := s.Get("foo")
	if !ok || v != "two" {
		t.Fatalf("Get(foo) = %q, %v; want \"two\", true", v, ok)
	}
}

func TestEmptyValueIsNotMissing(t *testing.T) {
	s := New()
	s.Set("foo", "")

	v, ok := s.Get("foo")
	if !ok || v != "" {
		t.Fatalf("Get(foo) = %q, %v; want \"\", true", v, ok)
	}
}

func TestDelete(t *testing.T) {
	s := New()
	s.Set("foo", "bar")
	s.Delete("foo")

	if _, ok := s.Get("foo"); ok {
		t.Fatal("Get(foo) found the key after Delete")
	}
}

func TestDeleteMissingIsNoop(t *testing.T) {
	s := New()
	s.Delete("never-existed") // must not panic
}

// TestConcurrent hammers the store from many goroutines.
//
// Run it twice:
//
//	go test ./store            -> crashes with "concurrent map writes" (no mutex yet)
//	go test -race ./store      -> prints a detailed data race report
//
// After you add a sync.RWMutex to Store, both should pass.
func TestConcurrent(t *testing.T) {
	s := New()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key%d", j%10)
				s.Set(key, fmt.Sprintf("v%d", n))
				s.Get(key)
				if j%100 == 0 {
					s.Delete(key)
				}
			}
		}(i)
	}

	wg.Wait()
}
