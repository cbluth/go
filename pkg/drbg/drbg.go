package drbg

import (
	"crypto/sha512"
)

type (
	DRBG []byte
)

func New(seed []byte) *DRBG {
	d := DRBG(seed)
	for i := 0; i < 1024; i++ {
		d.hash()
	}
	return &d
}

func (d *DRBG) Read(b []byte) (int, error) {
	n := 0
	for n < len(b) {
		n += copy(b[n:], d.hash())
	}
	return n, nil
}

func (d *DRBG) hash() []byte {
	h := sha512.Sum512(*d)
	*d = h[:32]
	return h[32:]
}
