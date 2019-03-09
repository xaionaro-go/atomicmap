package atomicmap

import (
	"runtime"
	"sync/atomic"
	"time"
)

type isSet uint32

const (
	isSet_notSet = isSet(iota) // 0
	isSet_set
	isSet_setting
	isSet_updating
	isSet_removed
)

func (i *isSet) Load() isSet {
	return (isSet)(atomic.LoadUint32((*uint32)(i)))
}

func (i *isSet) Store(newValue isSet) {
	atomic.StoreUint32((*uint32)(i), uint32(newValue))
}

type storageItem struct {
	// All the stuff what we need to grow. This variables are not connected
	// It's just all (unrelated) slices we need united into one to decrease
	// the number of memory allocations

	mapSlot mapSlot
	// ...other variables here...
}

type mapSlot struct {
	isSet        isSet
	readersCount int32
	hashValue    uint64
	slid         uint64 // how much items were already busy so we were have to go forward
	key          Key
	bytesValue   []byte
	value        interface{}
	fastKey      uint64
	fastKeyType  uint8
}

func (slot *mapSlot) IsSet() isSet {
	return slot.isSet.Load()
}

/*func (slot *mapSlot) waitForIsSet() bool {
	switch slot.IsSet() {
	case isSet_set:
		return true
	case isSet_notSet, isSet_removed:
		return false
	}

	runtime.Gosched()
	for {
		switch slot.IsSet() {
		case isSet_set:
			return true
		case isSet_notSet, isSet_removed:
			return false
		}
		time.Sleep(lockSleepInterval)
	}
}*/

func (i *isSet) CompareAndSwap(oldV, newV isSet) bool {
	return atomic.CompareAndSwapUint32((*uint32)(i), uint32(oldV), uint32(newV))
}

func (slot *mapSlot) setIsUpdating() bool {
	if slot.isSet.CompareAndSwap(isSet_set, isSet_updating) {
		return true
	}

	runtime.Gosched()
	for !slot.isSet.CompareAndSwap(isSet_set, isSet_updating) {
		if slot.IsSet() == isSet_removed {
			return false
		}
		time.Sleep(lockSleepInterval)
	}
	return true
}

func (slot *mapSlot) waitForReadersOut() {
	if atomic.LoadInt32(&slot.readersCount) == 0 {
		return
	}

	runtime.Gosched()
	for atomic.LoadInt32(&slot.readersCount) != 0 {
		time.Sleep(lockSleepInterval)
	}
}

func (slot *mapSlot) increaseReaders() isSet {
	atomic.AddInt32(&slot.readersCount, 1)
	isSet := slot.IsSet()
	switch isSet {
	case isSet_set:
		return isSet
	case isSet_notSet, isSet_removed:
		atomic.AddInt32(&slot.readersCount, -1)
		return isSet
	default:
		atomic.AddInt32(&slot.readersCount, -1)
	}
	runtime.Gosched()
	for {
		atomic.AddInt32(&slot.readersCount, 1)
		isSet := slot.IsSet()
		switch isSet {
		case isSet_set:
			return isSet
		case isSet_notSet, isSet_removed:
			atomic.AddInt32(&slot.readersCount, -1)
			return isSet
		default:
			atomic.AddInt32(&slot.readersCount, -1)
			time.Sleep(lockSleepInterval)
		}
	}
	panic(`Shouldn't happen`)
	return isSet_notSet
}

func (slot *mapSlot) decreaseReaders() {
	atomic.AddInt32(&slot.readersCount, -1)
}

type storage struct {
	items []storageItem
}

func newStorage(size uint64) *storage {
	return &storage{
		items: make([]storageItem, size),
	}
}

func (stor *storage) copyOldItemsAfterGrowing(oldStorage *storage) {
	if oldStorage == nil {
		return
	}
	if len(oldStorage.items) == 0 {
		return
	}
	for i := 0; i < len(oldStorage.items); i++ {
		oldSlot := &oldStorage.items[i].mapSlot
		if oldSlot.isSet == isSet_notSet {
			continue
		}

		newIdxValue := stor.getIdx(oldSlot.hashValue)
		newSlot, _, slid := stor.findFreeSlot(newIdxValue)
		copySlot(newSlot, oldSlot)
		newSlot.slid = slid
	}
}

func (stor *storage) size() uint64 {
	if stor == nil {
		return 0
	}
	return uint64(len(stor.items))
}

func getIdxHashMask(size uint64) uint64 { // this function requires size to be a power of 2
	return size - 1 // example 01000000 -> 00111111
}

func (stor *storage) getIdx(hashValue uint64) uint64 {
	return hashValue & getIdxHashMask(stor.size())
}

func (stor *storage) findFreeSlot(idxValue uint64) (*mapSlot, uint64, uint64) {
	var slotCandidate *mapSlot
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slotCandidate = &stor.items[idxValue].mapSlot
		if slotCandidate.isSet == isSet_notSet {
			return slotCandidate, idxValue, slid
		}
		slid++
		idxValue++
		if idxValue >= stor.size() {
			idxValue = 0
		}
	}
}
