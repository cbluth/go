package lru

import (
	"crypto/rand"
	"math"
	"math/big"
	"testing"
)

// CORE
////////

func TestLRUCore(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k int, v int) {
		if k != v {
			t.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	l, err := NewcLRU(128, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 256; i++ {
		l.add(i, i)
	}
	if l.len() != 128 {
		t.Fatalf("bad len: %v", l.len())
	}

	if evictCounter != 128 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.keys() {
		if v, ok := l.get(k); !ok || v != k || v != i+128 {
			t.Fatalf("bad key: %v", k)
		}
	}
	for i := 0; i < 128; i++ {
		if _, ok := l.get(i); ok {
			t.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		if _, ok := l.get(i); !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		if ok := l.remove(i); !ok {
			t.Fatalf("should be contained")
		}
		if ok := l.remove(i); ok {
			t.Fatalf("should not be contained")
		}
		if _, ok := l.get(i); ok {
			t.Fatalf("should be deleted")
		}
	}

	l.get(192) // expect 192 to be last key in l.Keys()

	for i, k := range l.keys() {
		if (i < 63 && k != i+193) || (i == 63 && k != 192) {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.purge()
	if l.len() != 0 {
		t.Fatalf("bad len: %v", l.len())
	}
	if _, ok := l.get(200); ok {
		t.Fatalf("should contain nothing")
	}
}

func TestCLRU_GetOldest_RemoveOldest(t *testing.T) {
	l, err := NewcLRU[int, int](128, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for i := 0; i < 256; i++ {
		l.add(i, i)
	}
	k, _, ok := l.getOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k != 128 {
		t.Fatalf("bad: %v", k)
	}

	// k, _, ok = l.removeOldest()

	ent := l.evictList.back()
	if ent != nil {
		l.removeElement(ent)
		k, ok = ent.key, true
	}
	if !ok {
		t.Fatalf("missing")
	}
	if k != 128 {
		t.Fatalf("bad: %v", k)
	}

	// k, _, ok = l.removeOldest()
	ent = l.evictList.back()
	if ent != nil {
		l.removeElement(ent)
		k, ok = ent.key, true
	}
	if !ok {
		t.Fatalf("missing")
	}
	if k != 129 {
		t.Fatalf("bad: %v", k)
	}
}

// Test that Add returns true/false if an eviction occurred
func TestLRU_Add(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k int, v int) {
		evictCounter++
	}

	l, err := NewcLRU(1, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if l.add(1, 1) == true || evictCounter != 0 {
		t.Errorf("should not have an eviction")
	}
	if l.add(2, 2) == false || evictCounter != 1 {
		t.Errorf("should have an eviction")
	}
}

// Test that Contains doesn't update recent-ness
func TestLRU_Contains(t *testing.T) {
	l, err := NewcLRU[int, int](2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.add(1, 1)
	l.add(2, 2)
	if !l.contains(1) {
		t.Errorf("1 should be contained")
	}

	l.add(3, 3)
	if l.contains(1) {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// Test that Peek doesn't update recent-ness
func TestLRU_Peek(t *testing.T) {
	l, err := NewcLRU[int, int](2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.add(1, 1)
	l.add(2, 2)
	if v, ok := l.peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.add(3, 3)
	if l.contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}
}

// Test that Resize can upsize and downsize
func TestLRU_Resize(t *testing.T) {
	onEvictCounter := 0
	onEvicted := func(k int, v int) {
		onEvictCounter++
	}
	l, err := NewcLRU(2, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Downsize
	l.add(1, 1)
	l.add(2, 2)
	evicted := l.resize(1)
	if evicted != 1 {
		t.Errorf("1 element should have been evicted: %v", evicted)
	}
	if onEvictCounter != 1 {
		t.Errorf("onEvicted should have been called 1 time: %v", onEvictCounter)
	}

	l.add(3, 3)
	if l.contains(1) {
		t.Errorf("Element 1 should have been evicted")
	}

	// Upsize
	evicted = l.resize(2)
	if evicted != 0 {
		t.Errorf("0 elements should have been evicted: %v", evicted)
	}

	l.add(4, 4)
	if !l.contains(3) || !l.contains(4) {
		t.Errorf("Cache should have contained 2 elements")
	}
}

// MODULE
////////////////////////////////////////////////////////////////

func BenchmarkLRU_Rand(b *testing.B) {
	l, err := New[int64, int64](8192, nil)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = getRand(b) % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Add(trace[i], trace[i])
		} else {
			if _, ok := l.Get(trace[i]); ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(hit+miss))
}

func BenchmarkLRU_Freq(b *testing.B) {
	l, err := New[int64, int64](8192, nil)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = getRand(b) % 16384
		} else {
			trace[i] = getRand(b) % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i])
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		if _, ok := l.Get(trace[i]); ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(hit+miss))
}

func TestLRU(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k int, v int) {
		if k != v {
			t.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	l, err := New(128, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	if l.Len() != 128 {
		t.Fatalf("bad len: %v", l.Len())
	}

	if evictCounter != 128 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || v != k || v != i+128 {
			t.Fatalf("bad key: %v", k)
		}
	}
	for i := 0; i < 128; i++ {
		if _, ok := l.Get(i); ok {
			t.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		if _, ok := l.Get(i); !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		l.Remove(i)
		if _, ok := l.Get(i); ok {
			t.Fatalf("should be deleted")
		}
	}

	l.Get(192) // expect 192 to be last key in l.Keys()

	for i, k := range l.Keys() {
		if (i < 63 && k != i+193) || (i == 63 && k != 192) {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		t.Fatalf("bad len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		t.Fatalf("should contain nothing")
	}
}

// test that Add returns true/false if an eviction occurred
func TestLRUAdd(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k int, v int) {
		evictCounter++
	}

	l, err := New(1, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if l.Add(1, 1) == true || evictCounter != 0 {
		t.Errorf("should not have an eviction")
	}
	if l.Add(2, 2) == false || evictCounter != 1 {
		t.Errorf("should have an eviction")
	}
}

// test that Contains doesn't update recent-ness
func TestLRUContains(t *testing.T) {
	l, err := New[int, int](2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	if !l.Contains(1) {
		t.Errorf("1 should be contained")
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// test that ContainsOrAdd doesn't update recent-ness
func TestLRUContainsOrAdd(t *testing.T) {
	l, err := New[int, int](2, func(key, value int) {})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	contains, evict := l.ContainsOrAdd(1, 1)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if evict {
		t.Errorf("nothing should be evicted here")
	}

	l.Add(3, 3)
	contains, evict = l.ContainsOrAdd(1, 1)
	if contains {
		t.Errorf("1 should not have been contained")
	}
	if !evict {
		t.Errorf("an eviction should have occurred")
	}
	if !l.Contains(1) {
		t.Errorf("now 1 should be contained")
	}
}

// test that PeekOrAdd doesn't update recent-ness
func TestLRUPeekOrAdd(t *testing.T) {
	l, err := New[int, int](2, func(key, value int) {})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	previous, contains, evict := l.PeekOrAdd(1, 1)
	if !contains {
		t.Errorf("1 should be contained")
	}
	if evict {
		t.Errorf("nothing should be evicted here")
	}
	if previous != 1 {
		t.Errorf("previous is not equal to 1")
	}

	l.Add(3, 3)
	_, contains, evict = l.PeekOrAdd(1, 1)
	if contains {
		t.Errorf("1 should not have been contained")
	}
	if !evict {
		t.Errorf("an eviction should have occurred")
	}
	if !l.Contains(1) {
		t.Errorf("now 1 should be contained")
	}
}

// test that Peek doesn't update recent-ness
func TestLRUPeek(t *testing.T) {
	l, err := New[int, int](2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	if v, ok := l.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}
}

// test that Resize can upsize and downsize
func TestLRUResize(t *testing.T) {
	onEvictCounter := 0
	onEvicted := func(k int, v int) {
		onEvictCounter++
	}
	l, err := New(2, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Downsize
	l.Add(1, 1)
	l.Add(2, 2)
	evicted := l.Resize(1)
	if evicted != 1 {
		t.Errorf("1 element should have been evicted: %v", evicted)
	}
	if onEvictCounter != 1 {
		t.Errorf("onEvicted should have been called 1 time: %v", onEvictCounter)
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("Element 1 should have been evicted")
	}

	// Upsize
	evicted = l.Resize(2)
	if evicted != 0 {
		t.Errorf("0 elements should have been evicted: %v", evicted)
	}

	l.Add(4, 4)
	if !l.Contains(3) || !l.Contains(4) {
		t.Errorf("Cache should have contained 2 elements")
	}
}

func TestLRU_GetOldest_RemoveOldest(t *testing.T) {
	l, err := New[int, int](128, func(key, value int) {})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	k, _, ok := l.GetOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k != 128 {
		t.Fatalf("bad: %v", k)
	}

	k, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k != 128 {
		t.Fatalf("bad: %v", k)
	}

	k, _, ok = l.RemoveOldest()
	// ent = l.evictList.back()
	// if ent != nil {
	// 	l.removeElement(ent)
	// 	k, ok = ent.key, true
	// }
	if !ok {
		t.Fatalf("missing")
	}
	if k != 129 {
		t.Fatalf("bad: %v", k)
	}
}

func getRand(tb testing.TB) int64 {
	out, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		tb.Fatal(err)
	}
	return out.Int64()
}


