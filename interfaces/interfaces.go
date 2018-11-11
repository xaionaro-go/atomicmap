package interfaces

type Key interface{}

type Map interface {
	Set(key Key, value interface{}) error
	Get(key Key) (value interface{}, err error)
	Unset(key Key) error
	Len() int
	ToSTDMap() map[Key]interface{}
	FromSTDMap(map[Key]interface{})
}
