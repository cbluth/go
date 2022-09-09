package drbg

import (
	"crypto/sha512"
)

type (
	DRBG struct {
		in, out []byte
	}
)

func New(seed []byte) *DRBG {
	d := &DRBG{
		in:  seed,
		out: []byte{},
	}
	for i := 0; i < 4096; i++ {
		d.out = d.hash()
	}
	return d
}

func (d *DRBG) Read(b []byte) (int, error) {
	n := 0
	for n < len(b) {
		n += copy(b[n:], d.hash())
	}
	return n, nil
}

func (d *DRBG) hash() []byte {
	h := sha512.Sum512(d.in)
	d.in =  h[:32]
	return h[32:]
}
