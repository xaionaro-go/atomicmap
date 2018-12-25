//go:generate benchmarkCodeGen

package builtinMap

import (
	e "errors"

	"github.com/xaionaro-go/atomicmap/errors"
	"github.com/xaionaro-go/atomicmap/hasher"
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

var (
	ErrNotImplemented = e.New("not implemented")
)

func NewWithArgs(blockSize uint64, customHasher hasher.Hasher) I.Map {
	return &builtinMap{}
}

// If you're going to forbid unhashable keys:
type builtinMap map[I.Key]interface{}

func (m builtinMap) Set(key I.Key, value interface{}) error {
	m[key] = value
	return nil
}
func (m builtinMap) Get(key I.Key) (interface{}, error) {
	value, ok := m[key]
	if !ok {
		return nil, errors.NotFound
	}
	return value, nil
}
func (m builtinMap) Unset(key I.Key) error {
	delete(m, key)
	return nil
}
func (m *builtinMap) Reset() {
	*m = builtinMap{}
}
func (m builtinMap) Hash(I.Key) uint64 {
	return 0
}
func (m *builtinMap) CheckConsistency() error {
	return nil
}
func (m *builtinMap) FromSTDMap(in map[I.Key]interface{}) {
	*m = in
}
func (m *builtinMap) ToSTDMap() map[I.Key]interface{} {
	return *m
}
func (m *builtinMap) GetByBytes(key []byte) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *builtinMap) Keys() []interface{} {
	return nil
}
func (m *builtinMap) Len() int {
	return len(*m)
}
