package li

import (
	"hash"

	"golang.org/x/crypto/blake2b"
)

type HashSum [32]byte

func NewHash() hash.Hash {
	h, _ := blake2b.New256(nil)
	return h
}

func init() {
	h := NewHash()
	if h.Size() != 32 {
		panic("bad hash size assumption")
	}
}
