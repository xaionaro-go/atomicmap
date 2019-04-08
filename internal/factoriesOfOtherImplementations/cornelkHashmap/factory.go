//go:generate benchmarkCodeGen

package cornelkHashmap

import (
	e "errors"

	"github.com/cornelk/hashmap"

	"github.com/xaionaro-go/atomicmap/errors"
	I "github.com/xaionaro-go/atomicmap/interfaces"
)

var (
	ErrNotImplemented = e.New("not implemented")
)

func New() I.Map {
	return &hashmapWrapper{}
}
func NewWithArgs(blockSize uint64) I.Map {
	return New()
}

type hashmapWrapper struct {
	hashmap.HashMap
}

func (m *hashmapWrapper) Get(key I.Key) (interface{}, error) {
	var err error
	v, ok := m.HashMap.Get(key)
	if !ok {
		err = errors.NotFound
	}
	return v, err
}

func (m *hashmapWrapper) FromSTDMap(map[I.Key]interface{}) {
}
func (m *hashmapWrapper) ToSTDMap() map[I.Key]interface{} {
	return nil
}
func (m *hashmapWrapper) Set(key I.Key, value interface{}) error {
	m.HashMap.Set(key, value)
	return nil
}
func (m *hashmapWrapper) SetBytesByBytes(k, v []byte) error {
	return ErrNotImplemented
}
func (m *hashmapWrapper) GetByBytes(key []byte) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *hashmapWrapper) GetByUint64(key uint64) (value interface{}, err error) {
	return nil, ErrNotImplemented
}
func (m *hashmapWrapper) Unset(key I.Key) error {
	m.HashMap.Del(key)
	return nil
}
func (m *hashmapWrapper) LockUnset(key I.Key) error {
	return m.Unset(key)
}
func (m *hashmapWrapper) Len() int {
	return -1
}
func (m *hashmapWrapper) Keys() []interface{} {
	return nil
}
func (m *hashmapWrapper) SetForbidGrowing(bool) {}
