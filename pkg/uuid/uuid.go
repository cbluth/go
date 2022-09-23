package uuid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type (
	UUID interface {
		Bytes() []byte
		String() string
		Array() [16]byte
	}
	uuid [16]byte
)

// New makes a new UUID interface;
// if input len is 0, executed like uuid.New(), then it will generate a random uuid;
// if input len is 16, like uuid.New([16]byte{}...), then it will directly use the 16 bytes as the uuid;
// if input len is something other than 0 or 16, then the input seeds a new psudo-random uuid.
func New(e ...byte) (UUID, error) {
	u := &uuid{}
	switch len(e) {
	case 0:
		{
			_, err := rand.Read(u[:])
			if err != nil {
				return nil, err
			}
		}
	case 16:
		{
			copy(u[:], e)
		}
	default:
		{
			s := sha256.Sum256(e)
			for i := 0; i < 32; i++ {
				s = sha256.Sum256(s[:16])
			}
			copy(u[:], s[:16])
		}
	}
	return u, nil
}

func (u *uuid) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

func (u *uuid) Bytes() []byte {
	return u[:]
}

func (u *uuid) Array() [16]byte {
	return *u
}

func StringtoUUID(s string) (UUID, error) {
	if len(strings.Split(s, "-")) != 5 {
		return nil, fmt.Errorf("probably not a uuid: %s", s)
	}
	b, err := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	if err != nil {
		return nil, fmt.Errorf("probably not a uuid: %v", err)
	}
	if len(b) != 16 {
		return nil, fmt.Errorf("probably not a uuid: %s", s)
	}
	return New(b...)
}
