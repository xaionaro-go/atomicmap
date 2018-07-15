//go:generate benchmarkCodeGen

package stupidMap

import (
	"fmt"

	"git.dx.center/trafficstars/testJob0/internal/errors"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

func NewHashMap(blockSize int, fn func(blockSize int, key I.Key) int) I.HashMaper {
	return &stupidMap{}
}

// If you're going to permit unhashable keys:
type stupidMap map[string]interface{}
func convertKey(keyI I.Key) string { // I.Key is interface{} and it can even be by something unhashable. So we have to represent it as a string.
	key, ok := keyI.(string)
	if ok {
		return "s"+key
	}
	return fmt.Sprintf("i%v", keyI)
}

/* // If you're going to forbid unhashable keys:
type stupidMap map[I.Key]interface{}
func convertKey(keyI I.Key) interface{} {
	return keyI
}*/

func (m stupidMap) Set(key I.Key, value interface{}) error {
	m[convertKey(key)] = value
	return nil
}
func (m stupidMap) Get(key I.Key) (interface{}, error) {
	value, ok := m[convertKey(key)]
	if !ok {
		return nil, errors.NotFound
	}
	return value, nil
}
func (m stupidMap) Unset(key I.Key) error {
	keyString := convertKey(key)
	_, ok := m[keyString]
	if !ok {
		return errors.NotFound
	}
	delete(m, keyString)
	return nil
}
func (m stupidMap) Count() int {
	return len(m)
}
func (m *stupidMap) Reset() {
	*m = stupidMap{}
}
func (m stupidMap) Hash(I.Key) uint64 {
	return 0
}
func (m *stupidMap) CheckConsistency() error {
	return nil
}
