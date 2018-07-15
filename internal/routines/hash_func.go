package routines

import (
	"fmt"
	"math"

	"github.com/OneOfOne/xxhash"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

const (
	randomNumber = uint64(4735311918715544114)
)

func preHashString(in string) uint64 {
	return uint64(xxhash.ChecksumString32(in))
}

func preHashPointer(in interface{}) uint64 {
	panic("not implemented")
}

func preHash(keyI I.Key) (value uint64, typeId uint8) {
	switch key := keyI.(type) {
	case bool:
		if key {
			return 1, 0
		} else {
			return 0, 0
		}
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
		return uint64(key), 6
	case int32:
		return uint64(key), 7
	case uint32:
		return uint64(key), 8
	case int64:
		return uint64(key), 9
	case uint64:
		return uint64(key), 10
	case float32:
		return uint64(math.Float32bits(key)), 11
	case float64:
		return uint64(math.Float64bits(key)), 12
	case complex64:
		return uint64(math.Float32bits(real(key)) ^ math.Float32bits(imag(key))), 13
	case complex128:
		return uint64(math.Float64bits(real(key)) ^ math.Float64bits(imag(key))), 14
	case *bool:
		return preHashPointer(key), 32
	case *string:
		return preHashPointer(key), 33
	case *int:
		return preHashPointer(key), 34
	case *uint:
		return preHashPointer(key), 35
	case *int8:
		return preHashPointer(key), 36
	case *uint8:
		return preHashPointer(key), 37
	case *int16:
		return preHashPointer(key), 38
	case *uint16:
		return preHashPointer(key), 39
	case *int32:
		return preHashPointer(key), 40
	case *uint32:
		return preHashPointer(key), 41
	case *int64:
		return preHashPointer(key), 42
	case *uint64:
		return preHashPointer(key), 43
	case *float32:
		return preHashPointer(key), 44
	case *float64:
		return preHashPointer(key), 45
	case *complex64:
		return preHashPointer(key), 46
	case *complex128:
		return preHashPointer(key), 47
	default:
		return preHashString(fmt.Sprintf("%v", key)), 63
	}
}

func HashFunc(blockSize int, key I.Key) int {
	preHashed, typeId := preHash(key)
	typeXorer := ((randomNumber << typeId) | (randomNumber >> (64 - typeId)))
	fullHash := preHashed ^ typeXorer
	hash := int(fullHash % uint64(blockSize))
	for fullHash > uint64(blockSize) {
		fullHash /= uint64(blockSize)
		hash ^= int(fullHash % uint64(blockSize))
	}
	return hash
}
