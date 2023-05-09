package lru

import (
	"errors"
	"sync"
)

const (
	// defaultEvictedBufferSize defines the default buffer size to store evicted key/val
	defaultEvictedBufferSize = 16
)

// LRU is the interface for simple LRU cache.
type LRU[K comparable, V any] interface {
	// Add a key/value, bool if evicted
	Add(key K, value V) bool
	// Get key/value, bool if found
	Get(key K) (value V, ok bool)
	// Contains checks for key
	Contains(key K) (ok bool)
	// Returns key's value without updating the "recently used"-ness of the key.
	Peek(key K) (value V, ok bool)
	// Removes a key from the cache.
	Remove(key K) bool
	// Removes the oldest entry from cache.
	RemoveOldest() (K, V, bool)
	// Returns the oldest entry from the cache. #key, value, isFound
	GetOldest() (K, V, bool)
	// Returns a slice of the keys in the cache, from oldest to newest.
	Keys() []K
	// Returns the number of items in the cache.
	Len() int
	// Clears all cache entries.
	Purge()
	// Resizes cache, returning number evicted
	Resize(int) int
	// ContainsOrAdd
	ContainsOrAdd(K, V) (bool, bool)
	// PeekOrAdd
	PeekOrAdd(K, V) (V, bool, bool)
}

// Cache is a thread-safe fixed size LRU cache.
type lru[K comparable, V any] struct {
	lru         *cLRU[K, V]
	evictedKeys []K
	evictedVals []V
	onEvictedCB func(k K, v V)
	sync.RWMutex
}

// New constructs a fixed size cache with the given eviction
// callback.
func New[K comparable, V any](size int, onEvicted func(key K, value V)) (LRU[K, V], error) {
	// create a cache with default settings
	c := &lru[K, V]{
		onEvictedCB: onEvicted,
	}
	if onEvicted != nil {
		c.initEvictBuffers()
		onEvicted = c.onEvicted
	}
	err := (error)(nil)
	c.lru, err = NewcLRU(size, onEvicted)
	return c, err
}

func (c *lru[K, V]) initEvictBuffers() {
	c.evictedKeys = make([]K, 0, defaultEvictedBufferSize)
	c.evictedVals = make([]V, 0, defaultEvictedBufferSize)
}

// onEvicted save evicted key/val and sent in externally registered callback
// outside of critical section
func (c *lru[K, V]) onEvicted(k K, v V) {
	c.evictedKeys = append(c.evictedKeys, k)
	c.evictedVals = append(c.evictedVals, v)
}

// Purge is used to completely clear the cache.
func (c *lru[K, V]) Purge() {
	var ks []K
	var vs []V
	c.Lock()
	c.lru.purge()
	if c.onEvictedCB != nil && len(c.evictedKeys) > 0 {
		ks, vs = c.evictedKeys, c.evictedVals
		c.initEvictBuffers()
	}
	c.Unlock()
	// invoke callback outside of critical section
	if c.onEvictedCB != nil {
		for i := 0; i < len(ks); i++ {
			c.onEvictedCB(ks[i], vs[i])
		}
	}
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (c *lru[K, V]) Add(key K, value V) (evicted bool) {
	var k K
	var v V
	c.Lock()
	evicted = c.lru.add(key, value)
	if c.onEvictedCB != nil && evicted {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.Unlock()
	if c.onEvictedCB != nil && evicted {
		c.onEvictedCB(k, v)
	}
	return
}

// Get looks up a key's value from the cache.
func (c *lru[K, V]) Get(key K) (value V, ok bool) {
	c.Lock()
	value, ok = c.lru.get(key)
	c.Unlock()
	return value, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *lru[K, V]) Contains(key K) bool {
	c.RLock()
	containKey := c.lru.contains(key)
	c.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *lru[K, V]) Peek(key K) (value V, ok bool) {
	c.RLock()
	value, ok = c.lru.peek(key)
	c.RUnlock()
	return value, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *lru[K, V]) ContainsOrAdd(key K, value V) (ok, evicted bool) {
	var k K
	var v V
	c.Lock()
	if c.lru.contains(key) {
		c.Unlock()
		return true, false
	}
	evicted = c.lru.add(key, value)
	if c.onEvictedCB != nil && evicted {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.Unlock()
	if c.onEvictedCB != nil && evicted {
		c.onEvictedCB(k, v)
	}
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *lru[K, V]) PeekOrAdd(key K, value V) (previous V, ok, evicted bool) {
	var k K
	var v V
	c.Lock()
	previous, ok = c.lru.peek(key)
	if ok {
		c.Unlock()
		return previous, true, false
	}
	evicted = c.lru.add(key, value)
	if c.onEvictedCB != nil && evicted {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.Unlock()
	if c.onEvictedCB != nil && evicted {
		c.onEvictedCB(k, v)
	}
	return
}

// Remove removes the provided key from the cache.
func (c *lru[K, V]) Remove(key K) (present bool) {
	var k K
	var v V
	c.Lock()
	present = c.lru.remove(key)
	if c.onEvictedCB != nil && present {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.Unlock()
	if c.onEvictedCB != nil && present {
		c.onEvictedCB(k, v)
	}
	return
}

// Resize changes the cache size.
func (c *lru[K, V]) Resize(size int) (evicted int) {
	var ks []K
	var vs []V
	c.Lock()
	evicted = c.lru.resize(size)
	if c.onEvictedCB != nil && evicted > 0 {
		ks, vs = c.evictedKeys, c.evictedVals
		c.initEvictBuffers()
	}
	c.Unlock()
	if c.onEvictedCB != nil && evicted > 0 {
		for i := 0; i < len(ks); i++ {
			c.onEvictedCB(ks[i], vs[i])
		}
	}
	return evicted
}

// RemoveOldest removes the oldest item from the cache.
func (c *lru[K, V]) RemoveOldest() (key K, value V, ok bool) {
	var k K
	var v V
	c.Lock()
	// key, value, ok = c.lru.removeOldest()
	ent := c.lru.evictList.back()
	if ent != nil {
		c.lru.removeElement(ent)
		key, value, ok = ent.key, ent.value, true
	}
	if c.onEvictedCB != nil && ok {
		k, v = c.evictedKeys[0], c.evictedVals[0]
		c.evictedKeys, c.evictedVals = c.evictedKeys[:0], c.evictedVals[:0]
	}
	c.Unlock()
	if c.onEvictedCB != nil && ok {
		c.onEvictedCB(k, v)
	}
	return
}

// GetOldest returns the oldest entry
func (c *lru[K, V]) GetOldest() (key K, value V, ok bool) {
	c.RLock()
	key, value, ok = c.lru.getOldest()
	c.RUnlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *lru[K, V]) Keys() []K {
	c.RLock()
	keys := c.lru.keys()
	c.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *lru[K, V]) Len() int {
	c.RLock()
	length := c.lru.len()
	c.RUnlock()
	return length
}

// CORE
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[K comparable, V any] func(key K, value V)

// cLRU implements a non-thread safe fixed size LRU cache
type cLRU[K comparable, V any] struct {
	size      int
	evictList *lruList[K, V]
	items     map[K]*entry[K, V]
	onEvict   EvictCallback[K, V]
}

// NewcLRU constructs an LRU of the given size
func NewcLRU[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*cLRU[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	c := &cLRU[K, V]{
		size:      size,
		evictList: newList[K, V](),
		items:     make(map[K]*entry[K, V]),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *cLRU[K, V]) purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.value)
		}
		delete(c.items, k)
	}
	c.evictList.init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *cLRU[K, V]) add(key K, value V) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.moveToFront(ent)
		ent.value = value
		return false
	}

	// Add new item
	ent := c.evictList.pushFront(key, value)
	c.items[key] = ent

	evict := c.evictList.length() > c.size
	// Verify size not exceeded
	if evict {
		if ent := c.evictList.back(); ent != nil {
			c.removeElement(ent)
		}
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *cLRU[K, V]) get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.moveToFront(ent)
		return ent.value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *cLRU[K, V]) contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *cLRU[K, V]) peek(key K) (value V, ok bool) {
	var ent *entry[K, V]
	if ent, ok = c.items[key]; ok {
		return ent.value, true
	}
	return
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *cLRU[K, V]) remove(key K) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}


// GetOldest returns the oldest entry
func (c *cLRU[K, V]) getOldest() (key K, value V, ok bool) {
	if ent := c.evictList.back(); ent != nil {
		return ent.key, ent.value, true
	}
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *cLRU[K, V]) keys() []K {
	keys := make([]K, c.evictList.length())
	i := 0
	for ent := c.evictList.back(); ent != nil; ent = ent.prevEntry() {
		keys[i] = ent.key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *cLRU[K, V]) len() int {
	return c.evictList.length()
}

// Resize changes the cache size.
func (c *cLRU[K, V]) resize(size int) (evicted int) {
	diff := c.len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		if ent := c.evictList.back(); ent != nil {
			c.removeElement(ent)
		}
	}
	c.size = size
	return diff
}

// // rremoveOldest removes the oldest item from the cache.
// func (c *cLRU[K, V]) rr2emoveOldest() {
// 	if ent := c.evictList.back(); ent != nil {
// 		c.removeElement(ent)
// 	}
// }

// // removeOldest removes the oldest item from the cache.
// func (c *cLRU[K, V]) oremoveOldest() (key K, value V, ok bool) {
// 	if ent := c.evictList.back(); ent != nil {
// 		c.removeElement(ent)
// 		return ent.key, ent.value, true
// 	}
// 	return
// }


// removeElement is used to remove a given list element from the cache
func (c *cLRU[K, V]) removeElement(e *entry[K, V]) {
	c.evictList.remove(e)
	delete(c.items, e.key)
	if c.onEvict != nil {
		c.onEvict(e.key, e.value)
	}
}

// entry is an LRU entry
type entry[K comparable, V any] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *entry[K, V]

	// The list to which this element belongs.
	list *lruList[K, V]

	// The LRU key of this element.
	key K

	// The value stored with this element.
	value V
}

// prevEntry returns the previous list element or nil.
func (e *entry[K, V]) prevEntry() *entry[K, V] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// lruList represents a doubly linked list.
// The zero value for lruList is an empty list ready to use.
type lruList[K comparable, V any] struct {
	root entry[K, V] // sentinel list element, only &root, root.prev, and root.next are used
	len  int         // current list length excluding (this) sentinel element
}

// init initializes or clears list l.
func (l *lruList[K, V]) init() *lruList[K, V] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// newList returns an initialized list.
func newList[K comparable, V any]() *lruList[K, V] { return new(lruList[K, V]).init() }

// length returns the number of elements of list l.
// The complexity is O(1).
func (l *lruList[K, V]) length() int { return l.len }

// back returns the last element of list l or nil if the list is empty.
func (l *lruList[K, V]) back() *entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List value.
func (l *lruList[K, V]) lazyInit() {
	if l.root.next == nil {
		l.init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *lruList[K, V]) insert(e, at *entry[K, V]) *entry[K, V] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *lruList[K, V]) insertValue(k K, v V, at *entry[K, V]) *entry[K, V] {
	return l.insert(&entry[K, V]{value: v, key: k}, at)
}

// remove removes e from its list, decrements l.len
func (l *lruList[K, V]) remove(e *entry[K, V]) V {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--

	return e.value
}

// move moves e to next to at.
func (l *lruList[K, V]) move(e, at *entry[K, V]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// pushFront inserts a new element e with value v at the front of list l and returns e.
func (l *lruList[K, V]) pushFront(k K, v V) *entry[K, V] {
	l.lazyInit()
	return l.insertValue(k, v, &l.root)
}

// moveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *lruList[K, V]) moveToFront(e *entry[K, V]) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}
