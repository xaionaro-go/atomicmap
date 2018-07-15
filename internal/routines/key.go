package routines

import (
	"fmt"
	"reflect"

	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

func IsEqualKey(keyA, keyB I.Key) bool {
	if reflect.TypeOf(keyA).Kind() != reflect.TypeOf(keyB).Kind() {
		return false
	}

	switch keyA.(type) {
	case bool:
		return keyA.(bool) == keyB.(bool)
	case string:
		return keyA.(string) == keyB.(string)
	case int:
		return keyA.(int) == keyB.(int)
	case uint:
		return keyA.(uint) == keyB.(uint)
	case int8:
		return keyA.(int8) == keyB.(int8)
	case uint8:
		return keyA.(uint8) == keyB.(uint8)
	case int16:
		return keyA.(int16) == keyB.(int16)
	case uint16:
		return keyA.(uint16) == keyB.(uint16)
	case int32:
		return keyA.(int32) == keyB.(int32)
	case uint32:
		return keyA.(uint32) == keyB.(uint32)
	case int64:
		return keyA.(int64) == keyB.(int64)
	case uint64:
		return keyA.(uint64) == keyB.(uint64)
	case float32:
		return keyA.(float32) == keyB.(float32)
	case float64:
		return keyA.(float64) == keyB.(float64)
	case complex64:
		return keyA.(complex64) == keyB.(complex64)
	case complex128:
		return keyA.(complex128) == keyB.(complex128)
	case *bool:
		return keyA.(*bool) == keyB.(*bool)
	case *string:
		return keyA.(*string) == keyB.(*string)
	case *int:
		return keyA.(*int) == keyB.(*int)
	case *uint:
		return keyA.(*uint) == keyB.(*uint)
	case *int8:
		return keyA.(*int8) == keyB.(*int8)
	case *uint8:
		return keyA.(*uint8) == keyB.(*uint8)
	case *int16:
		return keyA.(*int16) == keyB.(*int16)
	case *uint16:
		return keyA.(*uint16) == keyB.(*uint16)
	case *int32:
		return keyA.(*int32) == keyB.(*int32)
	case *uint32:
		return keyA.(*uint32) == keyB.(*uint32)
	case *int64:
		return keyA.(*int64) == keyB.(*int64)
	case *uint64:
		return keyA.(*uint64) == keyB.(*uint64)
	case *float32:
		return keyA.(*float32) == keyB.(*float32)
	case *float64:
		return keyA.(*float64) == keyB.(*float64)
	case *complex64:
		return keyA.(*complex64) == keyB.(*complex64)
	case *complex128:
		return keyA.(*complex128) == keyB.(*complex128)
	default:
		return fmt.Sprintf("%v", keyA) == fmt.Sprintf("%v", keyB)
	}
}
