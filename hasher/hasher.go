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

func (h *hasher) CompleteHash(keyPreHash uint64, keyTypeID uint8) uint64 {
	return CompleteHash(keyPreHash, keyTypeID)
}
func (h *hasher) CompressHash(blockSize uint64, fullHash uint64) uint64 {
	return CompressHash(blockSize, fullHash)
}
func (h *hasher) Hash(key interface{}) uint64 {
	return hash(key)
}
func (h *hasher) PreHashToKey(preHash uint64, typeID uint8) interface{} {
	return PreHashToKey(preHash, typeID)
}
func (h *hasher) PreHashToBytes(preHash uint64) []byte {
	return PreHashToBytes(preHash)
}
