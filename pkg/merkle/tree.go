package merkle

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"

	// "log"
	"sort"
	"sync"
)

const (
	KeySize = sha256.Size
)

var (
	_ Tree = &tree{}
)

type (
	Key  [KeySize]byte
	Tree interface {
		Add(Key)
		Remove(Key)
		Root() Key
		Len() int
		Depth() int
		Keys() []*Key
	}
	tree struct {
		depth int
		keys keys
		m    sync.RWMutex
		i    map[Key]struct{}
	}
	keys []*Key
)

func New() Tree {
	return &tree{
		keys: []*Key{},
		m:    sync.RWMutex{},
		i:    map[Key]struct{}{},
	}
}

func (ks keys) Less(i, j int) bool {
	ki, kj := Key(*ks[i]), Key(*ks[j])
	return bytes.Compare(ki[:], kj[:]) == -1
}

func (ks keys) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}

func (ks keys) Len() int { return len(ks) }

func (t *tree) Add(k Key) {
	t.m.Lock()
	defer t.m.Unlock()
	_, has := t.i[k]
	if has {
		return
	}
	t.i[k] = struct{}{}
	t.keys = append(t.keys, &k)
	sort.Sort(keys(t.keys))
}

func (t *tree) Remove(k Key) {
	t.m.Lock()
	defer t.m.Unlock()
	_, has := t.i[k]
	if !has {
		return
	}
	delete(t.i, k)
	for i, rk := range t.keys {
		if k == *rk {
			t.keys = append(t.keys[:i], t.keys[i+1:]...)
			return
		}
	}
}

func (t *tree) buildTree() Key {
	t.m.RLock()
	defer t.m.RUnlock()
	t.depth = 0
	return buildTree(t, t.keys)
}

func buildTree(t *tree, p keys) Key {
	l := len(p)
	if l == 0 {
		return Key{}
	}
	row := keys{}
	padded := false
	for i := 0; i < l; i += 2 {
		var left, right int = i, i + 1
		k := Key{}
		if i+1 == l {
			k = sha256.Sum256(p[left][:])
			padded = true
		} else {
			k = sha256.Sum256(
				append(p[left][:],p[right][:]...),
			)
		}
		row = append(row, (*Key)(&k))
		if padded {
			return k
		}
	}
	t.depth++
	return buildTree(t, row)
}

func (t *tree) Root() Key {
	return t.buildTree()
}

func (t *tree) Len() int {
	return len(t.keys)
}

func (t *tree) Depth() int {
	t.m.RLock()
	defer t.m.RUnlock()
	t.buildTree()
	return t.depth
}

func (t *tree) Keys() []*Key {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.keys
}

func (k Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}
