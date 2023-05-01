package ticket

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"log"
)

type (
	Machine interface {
		Issue([]byte) Ticket
		Validate(Ticket) []byte
	}
	machine struct {
		ed25519.PrivateKey
		ed25519.PublicKey
	}
	Ticket []byte
)

func NewMachine(privateKey ed25519.PrivateKey) Machine {
	return &machine{
		PrivateKey: privateKey,
		PublicKey:  privateKey.Public().(ed25519.PublicKey),
	}
}

func (m *machine) Issue(b []byte) Ticket {
	sig := ed25519.Sign(m.PrivateKey, b)
	msg := append(sig, b...)
	salt := [4]byte{}
	_, err := rand.Read(salt[:])
	if err != nil {
		log.Fatalf("crypto rand fail: %s", err)
	}
	secret := sha512.Sum512(append(salt[:], m.PrivateKey...))
	return append(salt[:], xor(secret[:], msg)...)
}

func (m *machine) Validate(t Ticket) []byte {
	salt := [4]byte{}
	copy(salt[:], t[:len(salt)])
	secret := sha512.Sum512(append(salt[:], m.PrivateKey...))
	b := xor(secret[:], t[len(salt):])
	if ed25519.Verify(m.PublicKey, b[ed25519.SignatureSize:], b[:ed25519.SignatureSize]) {
		return b[ed25519.SignatureSize:]
	}
	return nil
}

func xor(s, b []byte) []byte {
	n := 0
	for i := range b {
		b[i] ^= s[n]
		n++
		if n >= len(s) {
			n = 0
		}
	}
	return b
}

func (t Ticket) String() string {
	return base64.StdEncoding.EncodeToString(t)
}

func TicketFromString(s string) (Ticket, error) {
	t, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return t, nil

}
