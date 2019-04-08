package interfaces

type Key interface{}

type Map interface {
	Set(key Key, value interface{}) error
	SetBytesByBytes(key []byte, value []byte) error
	Get(key Key) (value interface{}, err error)
	GetByBytes(key []byte) (value interface{}, err error)
	GetByUint64(key uint64) (value interface{}, err error)
	Unset(key Key) error
	Len() int
	Keys() []interface{}
	ToSTDMap() map[Key]interface{}
	FromSTDMap(map[Key]interface{})
	SetForbidGrowing(bool)
}

type Hasher interface {
	PreHash(key interface{}) (uint64, uint8, bool)
	PreHashBytes(key []byte) (uint64, uint8, bool)
	PreHashUint64(key uint64) (uint64, uint8, bool)
	CompleteHash(preHash uint64, typeID uint8) uint64
	Hash(key interface{}) uint64
}
