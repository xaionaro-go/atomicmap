package hash

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

func preHashString(in string) uint64 {
	return uint64(xxhash.ChecksumString64(in))
}

func preHashPointer(in interface{}) uint64 {
	panic("not implemented")
}

func preHash(keyI I.Key) (value uint64, typeId uint8) {
	switch key := keyI.(type) {
	case string:
		return preHashString(key), 1
	case int:
		return uint64(key), 2
	case uint:
		return uint64(key), 3
	case int8:
		return uint64(key), 4
	case uint8:
		return uint64(key), 5
	case int16:
		return uint64(key), 6
	case uint16:
		return uint64(key), 7
	case int32:
		return uint64(key), 8
	case uint32:
		return uint64(key), 9
	case int64:
		return uint64(key), 10
	case uint64:
		return uint64(key), 11
	case float32:
		return uint64(math.Float32bits(key)), 12
	case float64:
		return uint64(math.Float64bits(key)), 13
	case complex64:
		return uint64(math.Float32bits(real(key)) ^ math.Float32bits(imag(key))), 14
	case complex128:
		return uint64(math.Float64bits(real(key)) ^ math.Float64bits(imag(key))), 15
	default:
		return preHashString(fmt.Sprintf("%v", key)), 63
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

func KeyHashFunc(blockSize uint64, key I.Key) uint64 {
	preHashed, typeId := preHash(key)
	if preHashed < blockSize {
		return preHashed % blockSize
	}
	typeXorer := bits.RotateLeft64(randomNumber, int(typeId))
	fullHash := preHashed ^ typeXorer
	return Uint64Hash(blockSize, fullHash)
}
