package interfaces

type Key interface{}

type Map interface {
	Set(key Key, value interface{}) error
	Get(key Key) (value interface{}, err error)
	Unset(key Key) error
	Count() int
	ToSTDMap() map[Key]interface{}
}
