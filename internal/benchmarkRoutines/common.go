package benchmarkRoutines

import (
	"encoding/binary"
	"math/rand"

	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

type mapFactoryFunc func(blockSize int, fn func(blockSize int, key I.Key) int) I.HashMaper
type hashFunc func(blockSize int, key I.Key) int

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
