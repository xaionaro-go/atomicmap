package hasher

import (
	"fmt"
	"math"
	"math/bits"
	//"sync/atomic"

	"github.com/OneOfOne/xxhash"
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

const (
	randomNumber           = uint64(4735311918715544114)
	knuthsMultiplicative8  = 179
	knuthsMultiplicative16 = 47351
	knuthsMultiplicative32 = 0x45d9f3b
	//knuthsMultiplicative64 = 1442695040888963407
)

var (
	counter = uint32(0)
)

func preHashString(in string) (uint64, uint8, bool) {
	if len(in) <= 8 {
		v := uint64(0)
		for i, c := range in {
			v += uint64(c) << (uint(i) << 3)
		}
		return v, 1, true
	}
	return xxhash.ChecksumString64(in), 1, false
}
func preHashBytes(in []byte) (uint64, uint8, bool) {
	if len(in) <= 8 {
		v := uint64(0)
		for i, c := range in {
			v += uint64(c) << (uint(i) << 3)
		}
		return v, 2, true
	}
	return xxhash.Checksum64(in), 2, false
}
func preHashUint64(in uint64) (uint64, uint8, bool) {
	return in, 12, true
}

func preHashPointer(in interface{}) uint64 {
	panic("not implemented")
}

func preHash(keyI I.Key) (value uint64, typeId uint8, isFull bool) {
	switch key := keyI.(type) {
	case string:
		return preHashString(key)
	case []byte:
		return preHashBytes(key)
	case int:
		return uint64(key), 3, true
	case uint:
		return uint64(key), 4, true
	case int8:
		return uint64(key), 5, true
	case uint8:
		return uint64(key), 6, true
	case int16:
		return uint64(key), 7, true
	case uint16:
		return uint64(key), 8, true
	case int32:
		return uint64(key), 9, true
	case uint32:
		return uint64(key), 10, true
	case int64:
		return uint64(key), 11, true
	case uint64:
		return preHashUint64(key)
	case float32:
		return uint64(math.Float32bits(key)), 13, true
	case float64:
		return uint64(math.Float64bits(key)), 14, true
	//case complex64:
	//	return uint64(math.Float32bits(real(key)) ^ math.Float32bits(imag(key))), 15
	case complex128:
		return uint64(math.Float64bits(real(key)) ^ math.Float64bits(imag(key))), 15, false
	default:
		preHash, _, isFullValue := preHashString(fmt.Sprintf("%v", key))
		return preHash, 63, isFullValue
	}
}

func Uint64Hash(blockSize uint64, key uint64) uint64 {
	subHash1 := uint32((key >> 32) ^ (key & 0xffffffff) ^ knuthsMultiplicative32)
	hash := uint64(subHash1 * knuthsMultiplicative32)
	if blockSize > (2 << 31) {
		return hash % blockSize
	}
	subHash2 := uint16((subHash1 >> 16) ^ (subHash1 & 0xffff) ^ knuthsMultiplicative16)
	hash ^= uint64(subHash2 * knuthsMultiplicative16)
	if blockSize > (2 << 15) {
		return hash % blockSize
	}
	subHash3 := uint8((subHash2 >> 8) ^ (subHash2 & 0xff) ^ knuthsMultiplicative8)
	hash ^= uint64(subHash3 * knuthsMultiplicative8)
	subHash4 := uint8((subHash3 >> 4) ^ (subHash3 & 0xf) ^ knuthsMultiplicative8)
	hash ^= uint64(subHash4 * knuthsMultiplicative8)
	return hash % blockSize
}

func completeHash(blockSize uint64, keyPreHash uint64, keyTypeID uint8) uint64 {
	if keyPreHash < blockSize {
		return keyPreHash
	}
	typeXorer := bits.RotateLeft64(randomNumber, int(keyTypeID))
	fullHash := keyPreHash ^ typeXorer
	return Uint64Hash(blockSize, fullHash)
}

func hash(blockSize uint64, key interface{}) uint64 {
	preHashValue, typeID, _ := preHash(key)
	return completeHash(blockSize, preHashValue, typeID)
}
