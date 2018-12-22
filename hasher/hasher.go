package hasher

import (
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

type Hasher = I.Hasher

type hasher struct {}

func New() Hasher {
	return &hasher{}
}

func (h *hasher) Hash(blockSize uint64, key interface{}) uint64 {
	return Hash(blockSize, key)
}

func (h *hasher) HashBytes(blockSize uint64, key []byte) uint64 {
	return HashBytes(blockSize, key)
}
