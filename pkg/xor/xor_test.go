package xor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	"github.com/cbluth/go/pkg/drbg"
)

var (
	myKey     = []byte("mysecret")
	myReader  = bytes.NewReader(myData)
	myData    = []byte("mydata;mydata;mydata;mydata;mydata;mydata;")
	nilerr    = "xor nil pointer"
	dataerr   = "missing data reader"
	secreterr = "missing secret reader"
)

// TestBytes .
func TestBytes(t *testing.T) {
	b, err := New(myKey).Bytes(myData)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	b, err = New(myKey).Bytes(b)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if !bytes.Equal(myData, b) {
		t.Errorf("expected same bytes")
	}
}

// TestBase64 .
func TestBase64(t *testing.T) {
	s, err := New(myKey).Base64(myData)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	s, err = New(myKey).Base64(b)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	b, err = base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if !bytes.Equal(myData, b) {
		t.Errorf("expected same bytes")
	}
}

// TestReader .
func TestReader(t *testing.T) {
	b, err := io.ReadAll(New(myKey).Reader(bytes.NewReader(myData)))
	if err != nil {
		t.Errorf("err: %v", err)
	}
	b, err = io.ReadAll(New(myKey).Reader(bytes.NewReader(b)))
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if !bytes.Equal(myData, b) {
		t.Errorf("expected same bytes")
	}
}

// TestNil .
func TestNil(t *testing.T) {
	x := (*xor)(nil)
	_, err := x.Bytes(myData)
	if err != nil && fmt.Sprintf("%v", err) != nilerr {
		t.Errorf("err: %v", err)
	}
	_, err = x.Base64(myData)
	if err != nil && fmt.Sprintf("%v", err) != nilerr {
		t.Errorf("err: %v", err)
	}
	_, err = io.ReadAll(x.Reader(myReader))
	if err != nil && fmt.Sprintf("%v", err) != nilerr {
		t.Errorf("err: %v", err)
	}
	x = &xor{}
	_, err = x.Bytes(myData)
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	_, err = x.Base64(myData)
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	_, err = io.ReadAll(x.Reader(myReader))
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	x = &xor{secret: drbg.New(myKey)}
	_, err = x.Bytes(nil)
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	_, err = x.Base64(nil)
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	_, err = io.ReadAll(x.Reader(nil))
	if err != nil && fmt.Sprintf("%v", err) != dataerr {
		t.Errorf("err: %v", err)
	}
	x = &xor{}
	_, err = x.Bytes(nil)
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	_, err = x.Base64(nil)
	if err != nil && fmt.Sprintf("%v", err) != secreterr {
		t.Errorf("err: %v", err)
	}
	_, err = io.ReadAll(x.Reader(nil))
	if err != nil && fmt.Sprintf("%v", err) != dataerr {
		t.Errorf("err: %v", err)
	}
}
