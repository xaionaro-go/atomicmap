package hasher

import (
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

type Hasher = I.Hasher

type hasher struct{}

func New() Hasher {
	return &hasher{}
}

func (h *hasher) PreHash(key interface{}) (uint64, uint8, bool) {
	return preHash(key)
}

func (h *hasher) PreHashBytes(key []byte) (uint64, uint8, bool) {
	return preHashBytes(key)
}

func (h *hasher) PreHashUint64(key uint64) (uint64, uint8, bool) {
	return preHashUint64(key)
}

func (h *hasher) CompleteHash(blockSize uint64, keyPreHash uint64, keyTypeID uint8) uint64 {
	return completeHash(blockSize, keyPreHash, keyTypeID)
}
func (h *hasher) Hash(blockSize uint64, key interface{}) uint64 {
	return hash(blockSize, key)
}
