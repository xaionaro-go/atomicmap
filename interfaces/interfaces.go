package interfaces

type Key interface{}

type Map interface {
	Set(key Key, value interface{}) error
	Get(key Key) (value interface{}, err error)
	GetByBytes(key []byte) (value interface{}, err error)
	GetByUint64(key uint64) (value interface{}, err error)
	LockUnset(key Key) error
	Unset(key Key) error
	Len() int
	Keys() []interface{}
	ToSTDMap() map[Key]interface{}
	FromSTDMap(map[Key]interface{})
}

type Hasher interface {
	Hash(blockSize uint64, key interface{}) uint64
	HashBytes(blockSize uint64, key []byte) uint64
	HashUint64(blockSize uint64, key uint64) uint64
}
