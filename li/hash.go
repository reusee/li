package li

import (
	"crypto/sha256"
	"hash"
)

type HashSum [32]byte

func NewHash() hash.Hash {
	return sha256.New()
}

func init() {
	h := NewHash()
	if h.Size() != 32 {
		panic("bad hash size assumption")
	}
}
