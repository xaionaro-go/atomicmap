package benchmarkRoutines

import (
	"encoding/binary"
	"math/rand"

	I "github.com/xaionaro-go/atomicmap/interfaces"
)

type mapFactoryFunc func(blockSize uint64, fn func(blockSize uint64, key I.Key) uint64) I.Map
type keyHashFunc func(blockSize uint64, key I.Key) uint64

type keyStruct struct {
	Key uint32
}

func generateKeys(keyAmount uint64, keyType string) []interface{} {
	resultMap := map[string]bool{}
	for uint64(len(resultMap)) < keyAmount {
		newKey := make([]byte, 4)
		rand.Read(newKey)
		resultMap[string(newKey)] = true
	}

	i := 0
	result := make([]interface{}, keyAmount, keyAmount)
	for newKey := range resultMap {
		newKeyInt := binary.LittleEndian.Uint32([]byte(newKey))
		switch keyType {
		case "int":
			result[i] = newKeyInt
		case "string":
			result[i] = newKey
		case "map":
			result[i] = map[uint32]uint32{newKeyInt: newKeyInt}
		case "slice":
			result[i] = []uint32{newKeyInt}
		case "struct":
			result[i] = keyStruct{Key: newKeyInt}
		default:
			panic("Unknown key type: " + keyType)
		}
		i++
	}
	return result
}
