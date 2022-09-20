package uuid

import (
	"crypto/rand"
	"fmt"
)
type (
	UUID interface {
		Bytes() [16]byte
		String() string
	}
	uuid [16]byte
)

func New() (UUID, error) {
	u := &uuid{}
	_, err := rand.Read(u[:])
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *uuid) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func (u *uuid) Bytes() [16]byte {
	return *u
}