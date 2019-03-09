package atomicmap

import (
	"runtime"
	"sync/atomic"
	"time"
	"github.com/xaionaro-go/atomicmap/hasher"
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
	isSet        isSet
	readersCount int32
	hashValue    uint64
	slid         uint64 // how much items were already busy so we were have to go forward (if previous items were removed)
	key          Key
	bytesValue   []byte
	value        interface{}
	fastKey      uint64
	fastKeyType  uint8
}

func (slot *storageItem) IsSet() isSet {
	return (*isSet)(&slot.isSet).Load()
}

func (slot *storageItem) IsSetCompareAndSwap(oldV, newV isSet) bool {
	return (*isSet)(&slot.isSet).CompareAndSwap(oldV, newV)
}

func (slot *storageItem) IsSetStore(newV isSet) {
	(*isSet)(&slot.isSet).Store(newV)
}

/*func (slot *storageItem) waitForIsSet() bool {
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

func (slot *storageItem) setIsUpdating() bool {
	if slot.IsSetCompareAndSwap(isSet_set, isSet_updating) {
		return true
	}

	runtime.Gosched()
	for !slot.IsSetCompareAndSwap(isSet_set, isSet_updating) {
		if slot.IsSet() == isSet_removed {
			return false
		}
		time.Sleep(lockSleepInterval)
	}
	return true
}

func (slot *storageItem) waitForReadersOut() {
	if atomic.LoadInt32(&slot.readersCount) == 0 {
		return
	}

	runtime.Gosched()
	for atomic.LoadInt32(&slot.readersCount) != 0 {
		time.Sleep(lockSleepInterval)
	}
}
func (slot *storageItem) increaseReadersStage0Sub0Sub0() {
	atomic.AddInt32(&slot.readersCount, 1)
}
func (slot *storageItem) increaseReadersStage0Sub0Sub1() isSet {
	return slot.IsSet()
}
func (slot *storageItem) increaseReadersStage0Sub0() isSet {
	slot.increaseReadersStage0Sub0Sub0()
	return slot.increaseReadersStage0Sub0Sub1()
}
func (slot *storageItem) increaseReadersStage0() isSet {
	isSetR := slot.increaseReadersStage0Sub0()
	switch isSetR {
	case isSet_set:
		return isSetR
	case isSet_notSet, isSet_removed:
		atomic.AddInt32(&slot.readersCount, -1)
		return isSetR
	default:
		atomic.AddInt32(&slot.readersCount, -1)
	}

	return isSet(10)
}
func (slot *storageItem) increaseReadersStage1() isSet {
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

func (slot *storageItem) increaseReaders() isSet {
	r := slot.increaseReadersStage0()
	if r == isSet(10) {
		r = slot.increaseReadersStage1()
	}
	return r
}

func (slot *storageItem) decreaseReaders() {
	atomic.AddInt32(&slot.readersCount, -1)
}

type storage struct {
	hasher           hasher.Hasher
	threadSafety bool
	items []storageItem
}

func newStorage(size uint64, hasher hasher.Hasher, threadSafety bool) *storage {
	stor := &storage{
		hasher: hasher,
		items: make([]storageItem, size),
		threadSafety: threadSafety,
	}

	return stor
}

func (stor *storage) getItem(idx uint64) *storageItem {
	return &stor.items[idx]
}

func (stor *storage) copyOldItemsAfterGrowing(oldStorage *storage) {
	if oldStorage == nil {
		return
	}
	if len(oldStorage.items) == 0 {
		return
	}
	for i := uint64(0); i < uint64(len(oldStorage.items)); i++ {
		oldSlot := oldStorage.getItem(i)
		if isSet(oldSlot.isSet) == isSet_notSet {
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

/*func getIdxHashMask(size uint64) uint64 { // this function requires size to be a power of 2
	return size - 1 // example 01000000 -> 00111111
}*/

func (stor *storage) getIdx(hashValue uint64) uint64 {
	return stor.hasher.CompressHash(stor.size(), hashValue)
}

func (stor *storage) findFreeSlot(idxValue uint64) (*storageItem, uint64, uint64) {
	var slotCandidate *storageItem
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slotCandidate = stor.getItem(idxValue)
		if isSet(slotCandidate.isSet) == isSet_notSet {
			return slotCandidate, idxValue, slid
		}
		slid++
		idxValue++
		if idxValue >= stor.size() {
			idxValue = 0
		}
	}
}

func (stor *storage) getByHashValue(preHashValue uint64, typeID uint8, preHashValueIsFull bool, isRightSlotFn func(*storageItem) bool) (interface{}, error) {
	hashValue := stor.hasher.CompleteHash(preHashValue, typeID)
	fastKey, fastKeyType := preHashValue, typeID
	idxValue := stor.getIdx(hashValue)

	for {
		slot := stor.getItem(idxValue)
		idxValue++
		if idxValue >= stor.size() {
			idxValue = 0
		}
		var isSetStatus isSet
		if stor.threadSafety {
			isSetStatus = slot.increaseReaders()
		} else {
			isSetStatus = slot.IsSet()
		}
		if isSetStatus == isSet_notSet {
			break
		}
		if isSetStatus == isSet_removed || slot.hashValue != hashValue {
			slot.decreaseReaders()
			continue
		}
		var isRightSlot bool
		if slot.fastKeyType != 0 || preHashValueIsFull {
			isRightSlot = slot.fastKey == fastKey && slot.fastKeyType == fastKeyType && preHashValueIsFull
		} else {
			isRightSlot = isRightSlotFn(slot)
		}

		if !isRightSlot {
			slot.decreaseReaders()
			continue
		}

		var value interface{}
		if slot.bytesValue != nil {
			value = slot.bytesValue
		} else {
			value = slot.value
		}
		slot.decreaseReaders()
		//m.decreaseConcurrency()
		return value, nil
	}

	//m.decreaseConcurrency()
	return nil, NotFound
}
