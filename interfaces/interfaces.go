package interfaces

type Key interface{}

type HashMaper interface {
	Set(key Key, value interface{}) error
	Get(key Key) (value interface{}, err error)
	Unset(key Key) error
	Count() int
}
