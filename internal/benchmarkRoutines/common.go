package benchmarkRoutines

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"math/rand"

	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

type mapFactoryFunc func(blockSize int, fn func(blockSize int, key I.Key) int) I.HashMaper

func getKeyBytes(key I.Key) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generateKeys(keyAmount uint64, keyIsString bool) []interface{} {
	resultMap := map[string]bool{}
	for uint64(len(resultMap)) < keyAmount {
		newKey := make([]byte, 4)
		rand.Read(newKey)
		resultMap[string(newKey)] = true
	}

	i := 0
	result := make([]interface{}, keyAmount, keyAmount)
	for newKey := range resultMap {
		if keyIsString {
			result[i] = newKey
		} else {
			result[i] = binary.LittleEndian.Uint32([]byte(newKey))
		}
		i++
	}
	return result
}
