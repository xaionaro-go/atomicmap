//go:generate benchmarkCodeGen

package builtinSyncMap

import (
	e "errors"
	"sync"

	"github.com/xaionaro-go/atomicmap/errors"
	"github.com/xaionaro-go/atomicmap/hasher"
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

var (
	ErrNotImplemented = e.New("not implemented")
)

func NewWithArgs(blockSize uint64, customHasher hasher.Hasher) I.Map {
	return &builtinSyncMap{}
}

// If you're going to forbid unhashable keys:
type builtinSyncMap struct {
	sync.Map
}

func (m *builtinSyncMap) Set(key I.Key, value interface{}) error {
	m.Map.Store(key, value)
	return nil
}
func (m *builtinSyncMap) SetBytesByBytes(k, v []byte) error {
	return ErrNotImplemented
}
func (m *builtinSyncMap) Get(key I.Key) (interface{}, error) {
	value, ok := m.Map.Load(key)
	if !ok {
		return value, errors.NotFound
	}
	return value, nil
}
func (m *builtinSyncMap) Unset(key I.Key) error {
	m.Map.Delete(key)
	return nil
}
func (m *builtinSyncMap) LockUnset(key I.Key) error {
	return m.Unset(key)
}
func (m *builtinSyncMap) Reset() {
	*m = builtinSyncMap{}
}
func (m *builtinSyncMap) Hash(I.Key) uint64 {
	return 0
}
func (m *builtinSyncMap) CheckConsistency() error {
	return nil
}
func (m *builtinSyncMap) FromSTDMap(in map[I.Key]interface{}) {
	return
}
func (m *builtinSyncMap) ToSTDMap() map[I.Key]interface{} {
	return nil
}
func (m *builtinSyncMap) GetByBytes(key []byte) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *builtinSyncMap) GetByUint64(key uint64) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *builtinSyncMap) Keys() []interface{} {
	return nil
}
func (m *builtinSyncMap) Len() int {
	return -1
}
func (m *builtinSyncMap) SetForbidGrowing(bool) {}
