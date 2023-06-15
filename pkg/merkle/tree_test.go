package merkle

import (
	"crypto/sha256"
	"encoding/base64"
	"log"

	// "log"
	// "log"
	"testing"
)

var (
	testKeys []Key = []Key{
		sha256.Sum256([]byte("k1")),  // arnx6499M4j0+dWG9m6Z/VQIDfLERvDlhmiwnAihbdA=
		sha256.Sum256([]byte("k2")),  // AV9+a8Wur0g3JAieklLME7UJUaa2lBJSJ2XP9NeAMG4=
		sha256.Sum256([]byte("k3")),  // L1BSyf0VsZoYxYTQE2NWgZhhPww06EQJ73k4cJoVnsI=
		sha256.Sum256([]byte("k4")),  //
		sha256.Sum256([]byte("k5")),  //
		sha256.Sum256([]byte("k6")),  //
		sha256.Sum256([]byte("k7")),  //
		sha256.Sum256([]byte("k8")),  //
		sha256.Sum256([]byte("k9")),  //
		sha256.Sum256([]byte("k10")), //
		sha256.Sum256([]byte("k11")), //
		sha256.Sum256([]byte("k12")), //
	}
	r0 = decodeBase64Panic("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	r1 Key = sha256.Sum256([]byte("k1"))
	r2 Key = sha256.Sum256(
		append(testKeys[0][:], testKeys[1][:]...),
	)
	rootKeys []Key = []Key{
		// decodeBase64Panic("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="),
		decodeBase64Panic("Zh7sxdGkMxeNjM+0tKEWeZlRBFjhZlzy93kFx/KkCmc="),
		decodeBase64Panic("q4g7ifVbu7TfWVZPYaNJtiM8XrcMMBuLhiTiOPrKw00="),
		decodeBase64Panic("8Z4eZVEFFhY+BYDU/hXMxx1kZW0oMVYC424VFUIEoVQ="),
		// decodeBase64Panic("weSQezz9BjmQTNJgKLsKMQU9AKP2noLOc55XpQfI8Uk="),
		// decodeBase64Panic("x+mSwqeUPGABcrICYs+mUy2Ule7/bI0W4qq5Rhxd/g4="),
		// decodeBase64Panic("kPV9EqQt2jfOb1j3KbtUKB4AJs5MOfxNsb9nnF7IX2k="),
		// decodeBase64Panic("Zh7sxdGkMxeNjM+0tKEWeZlRBFjhZlzy93kFx/KkCmc="),
		// decodeBase64Panic("Zh7sxdGkMxeNjM+0tKEWeZlRBFjhZlzy93kFx/KkCmc="),
		// decodeBase64Panic("Zh7sxdGkMxeNjM+0tKEWeZlRBFjhZlzy93kFx/KkCmc="),
		// decodeBase64Panic("Zh7sxdGkMxeNjM+0tKEWeZlRBFjhZlzy93kFx/KkCmc="),
		// decodeBase64Panic("Zh7sxdGkMxeNjM+0tKEWeZlRBFjhZlzy93kFx/KkCmc="),
	}
)

// TestAddRemove .
func TestAddRemove(t *testing.T) {
	m := New()
	if m.Len() != 0 {
		t.Errorf("expecting length 0")
	}
	// add
	for i, k := range testKeys {
		if m.Len() != i {
			t.Errorf("add: expecting length %v, is %v", i, m.Len())
		}
		m.Add(k)
		if m.Len() != i+1 {
			t.Errorf("add: expecting length %v, is %v", i+1, m.Len())
		}
		m.Add(k)
		if m.Len() != i+1 {
			t.Errorf("add: expecting length %v, is %v", i+1, m.Len())
		}
	}
	m.Keys()
	// remove
	for i, k := range testKeys {
		if m.Len() != len(testKeys)-i {
			t.Errorf("remove: expecting length %v, is %v", len(testKeys)-i, m.Len())
		}
		m.Remove(k)
		if m.Len() != len(testKeys)-i-1 {
			t.Errorf("remove: expecting length %v, is %v", len(testKeys)-i-1, m.Len())
		}
		m.Remove(k)
		if m.Len() != len(testKeys)-i-1 {
			t.Errorf("remove: expecting length %v, is %v", len(testKeys)-i-1, m.Len())
		}
	}
	if m.Len() != 0 {
		t.Errorf("expecting length 0")
	}
}

func TestRoots(t *testing.T) {
	m := New()
	r := m.Root()
	if r != r0 {
		t.Errorf("expecting root %v, is %v", r0.String(), r.String())
	}
	m.Add(testKeys[0])
	r = m.Root()
	if r != r1 {
		t.Errorf(": expecting root %v, is %v", r1.String(), r.String())
	}
	m.Add(testKeys[1])
	r = m.Root()
	if r != r2 {
		t.Errorf(": expecting root %v, is %v", r2.String(), r.String())
	}
	// for i, k := range testKeys {
	// 	m.Add(k)
	// 	r := m.Root()
	// 	if r != rootKeys[i] {
	// 		t.Errorf("%v: expecting root %v, is %v",i, rootKeys[i].String(), r.String())
	// 	}
	// }
	// m.Add(testKeys[0])
	// r = m.Root()
	// if r != r1 {
	// 	t.Errorf("expecting root %v, is %v", r1.String(), r.String())
	// }
}

func TestDepth(t *testing.T) {
	m := New()
	// 0 keys
	if i := 0 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
	// 1 key
	m.Add(testKeys[0])
	if i := 1 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
	// 2 keys
	m.Add(testKeys[1])
	if i := 2 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
	// 3 keys
	m.Add(testKeys[2])
	if i := 3 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
	// 4 keys
	m.Add(testKeys[3])
	if i := 3 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
	// 5 keys
	m.Add(testKeys[4])
	if i := 4 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
	// 6 keys
	m.Add(testKeys[5])
	if i := 4 ; m.Depth() != i {
		t.Errorf("expecting depth %v, is %v, len %v", i, m.Depth(), m.Len())
	}
}

func decodeBase64Panic(s string) Key {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Fatalln(err)
	}
	k := Key{}
	copy(k[:], b)
	return k
}

// func calculateK1K2Root() Key {
// 	k := sha256.Sum256(
// 		append(testKeys[0][:], testKeys[1][:]...),
// 	)
// }
