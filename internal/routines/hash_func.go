package routines

import (
	"fmt"
	"math"
	"math/bits"
	//"sync/atomic"

	I "git.dx.center/trafficstars/testJob0/task/interfaces"
	"github.com/OneOfOne/xxhash"
)

const (
	randomNumber         = uint64(4735311918715544114)
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

func HashFunc(blockSize int, key I.Key) int {
	preHashed, typeId := preHash(key)
	typeXorer := bits.RotateLeft64(randomNumber, int(typeId))
	fullHash := preHashed ^ typeXorer
	hash := uint64(0)
	subHash1 := (fullHash >> 32) ^ (fullHash & 0xffffffff) ^ knuthsMultiplicative32
	hash ^= uint64(uint32(subHash1) * knuthsMultiplicative32)
	subHash2 := (subHash1 >> 16) ^ (subHash1 & 0xffff) ^ knuthsMultiplicative16
	hash ^= uint64(uint16(subHash2) * knuthsMultiplicative16)
	subHash3 := (subHash2 >> 8) ^ (subHash2 & 0xff) ^ knuthsMultiplicative8
	hash ^= uint64(uint8(subHash3) * knuthsMultiplicative8)
	subHash4 := (subHash3 >> 4) ^ (subHash3 & 0x7) ^ knuthsMultiplicative8
	hash ^= uint64(uint8(subHash4) * knuthsMultiplicative8)
	return int(hash) % blockSize
}
