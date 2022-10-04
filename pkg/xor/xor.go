package xor

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/cbluth/go/pkg/drbg"
)

var (
	_         io.Reader = &xor{}
	errNil              = fmt.Errorf("xor nil pointer")
	errData             = fmt.Errorf("missing data reader")
	errSecret           = fmt.Errorf("missing secret reader")
)

type (
	XOR interface {
		Read([]byte) (int, error)
		Reader(io.Reader) io.Reader
		Bytes([]byte) ([]byte, error)
		Base64([]byte) (string, error)
	}
	xor struct {
		data   io.Reader
		secret io.Reader
	}
)

func New(seed []byte) XOR {
	s := sha512.Sum512(seed)
	return &xor{
		secret: drbg.New(s[:]),
	}
}

func (x *xor) Bytes(in []byte) ([]byte, error) {
	if x == nil {
		return nil, errNil
	}
	x.data = bytes.NewReader(in)
	return io.ReadAll(x)
}

func (x *xor) Base64(in []byte) (string, error) {
	if x == nil {
		return "", errNil
	}
	x.data = bytes.NewReader(in)
	b, err := io.ReadAll(x)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (x *xor) Read(b []byte) (int, error) {
	if x == nil {
		return 0, errNil
	}
	if x.data == nil {
		return 0, errData
	}
	if x.secret == nil {
		return 0, errSecret
	}
	n, err := x.data.Read(b)
	b = b[:n]
	if n > 0 {
		s := make([]byte, n)
		io.ReadFull(x.secret, s)
		for i := 0; i < n; i++ {
			b[i] ^= s[i]
		}
	}
	return n, err
}

func (x *xor) Reader(in io.Reader) io.Reader {
	if x == nil {
		return x
	}
	x.data = in
	return x
}
