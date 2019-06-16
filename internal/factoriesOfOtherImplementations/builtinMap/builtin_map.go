//go:generate benchmarkCodeGen

package builtinMap

import (
	e "errors"
	"sync"

	"github.com/xaionaro-go/atomicmap/errors"
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

var (
	ErrNotImplemented = e.New("not implemented")
)

func NewWithArgs(blockSize uint64) I.Map {
	return &builtinMap{
		m: make(map[I.Key]interface{}),
	}
}

// If you're going to forbid unhashable keys:
type builtinMap struct {
	sync.Mutex

	m map[I.Key]interface{}
}

func (m *builtinMap) Set(key I.Key, value interface{}) error {
	m.m[key] = value
	return nil
}
func (m *builtinMap) Swap(key I.Key, value interface{}) (interface{}, error) {
	oldValue := m.m[key]
	m.m[key] = value
	return oldValue, nil
}
func (m *builtinMap) SetBytesByBytes(k, v []byte) error {
	return ErrNotImplemented
}
func (m *builtinMap) Get(key I.Key) (interface{}, error) {
	value, ok := m.m[key]
	if !ok {
		return nil, errors.NotFound
	}
	return value, nil
}
func (m *builtinMap) Unset(key I.Key) error {
	delete(m.m, key)
	return nil
}
func (m *builtinMap) LockUnset(key I.Key) error {
	m.Lock()
	delete(m.m, key)
	m.Unlock()
	return nil
}
func (m *builtinMap) Reset() {
	*m = builtinMap{}
}
func (m *builtinMap) Hash(I.Key) uint64 {
	return 0
}
func (m *builtinMap) CheckConsistency() error {
	return nil
}
func (m *builtinMap) FromSTDMap(in map[I.Key]interface{}) {
	m.m = in
}
func (m *builtinMap) ToSTDMap() map[I.Key]interface{} {
	return m.m
}
func (m *builtinMap) GetByBytes(key []byte) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *builtinMap) GetByUint64(key uint64) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *builtinMap) Keys() []interface{} {
	return nil
}
func (m *builtinMap) Len() int {
	return len(m.m)
}
func (m *builtinMap) SetForbidGrowing(bool) {}
