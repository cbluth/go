package merkle

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"math"

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
		// depth int
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
	a, b := Key(*ks[i]), Key(*ks[j])
	return bytes.Compare(a[:], b[:]) == -1
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
	sort.Stable(t.keys)
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

func buildTree(p keys) Key {
	l := len(p)
	switch l {
	case 0:
		return Key{}
	case 1:
		return *p[0]
	}
	row := keys{}
	padded := l%2 != 0
	for i := 0 ; i < l ; i +=2 {
		pl, pr := i, i + 1
		k := Key{}
		if i+1 == l && padded {
			k = sha256.Sum256(p[pl][:])
		} else {
			k = sha256.Sum256(
				append(p[pl][:],p[pr][:]...),
			)
		}
		row = append(row, (*Key)(&k))
		switch l {
		case 1, 2:
			return k
		}
	}
	return buildTree(row)
}

// func (t *tree) buildTree() Key {
// 	t.m.Lock()
// 	defer t.m.Unlock()
// 	t.depth = 0
// 	return buildTree(t, t.keys)
// }

// func buildTree(t *tree, p keys) Key {
// 	l := len(p)
// 	if l == 0 {
// 		return Key{}
// 	}
// 	row := keys{}
// 	padded := false
// 	for i := 0; i < l; i += 2 {
// 		var left, right int = i, i + 1
// 		k := Key{}
// 		if i+1 == l {
// 			k = sha256.Sum256(p[left][:])
// 			padded = true
// 		} else {
// 			k = sha256.Sum256(
// 				append(p[left][:],p[right][:]...),
// 			)
// 		}
// 		row = append(row, (*Key)(&k))
// 		if padded {
// 			return k
// 		}
// 	}
// 	t.depth++
// 	return buildTree(t, row)
// }

func (t *tree) Root() Key {
	t.m.RLock()
	defer t.m.RUnlock()
	// return t.buildTree()
	return buildTree(t.keys)
}

func (t *tree) Len() int {
	t.m.RLock()
	defer t.m.RUnlock()
	return len(t.keys)
}

func (t *tree) Depth() int {
	t.m.RLock()
	defer t.m.RUnlock()
	l := len(t.keys)
	switch l {
	case 0, 1, 2:
		return l
	}
	return int(math.Ceil(math.Log2(float64(l)))) + 1
}

func (t *tree) Keys() []*Key {
	t.m.RLock()
	defer t.m.RUnlock()
	return t.keys
}

func (k Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}
