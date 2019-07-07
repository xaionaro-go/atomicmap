//go:generate benchmarkCodeGen

package atomicmap

import (
	"fmt"
	"github.com/trafficstars/go/src/math"
	"log"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/xaionaro-go/atomicmap/hasher"
	"github.com/xaionaro-go/spinlock"
)

const (
	growAtFullness    = 0.85
	maximalSize       = 1 << 32
	lockSleepInterval = 300 * time.Nanosecond
	defaultBlockSize  = 65536
)

var (
	threadSafe    = true
	forbidGrowing = false
)

type Map = *openAddressGrowingMap

func powerOfTwoGT(v uint64) uint64 {
	shiftedV := v
	for (v+1)^v < (v + 1) {
		shiftedV >>= 1
		v |= shiftedV
	}
	v++
	return v
}

func isPowerOfTwo(v uint64) bool {
	return (v-1)^v == v
}

func powerOfTwoGE(v uint64) uint64 {
	if isPowerOfTwo(v) {
		return v
	}
	return powerOfTwoGT(v)
}

func fixBlockSize(blockSize uint64) uint64 {
	// This functions fixes blockSize value to be a power of 2

	if blockSize <= 0 {
		log.Printf("Invalid block size: %v. Setting to %d\n", blockSize, defaultBlockSize)
		blockSize = defaultBlockSize
	}

	if (blockSize-1)^blockSize < blockSize {
		blockSize = powerOfTwoGT(blockSize)
		log.Printf("blockSize should be a power of 2 (1, 2, 4, 8, 16, ...). Setting to %v", blockSize)
	}

	return blockSize
}

func New() Map {
	return NewWithArgs(0)
}

// blockSize should be a power of 2 and should be greater than the maximal amount of elements you're planning to store. Keep in mind: the higher blockSize you'll set the longer initialization will be (and more memory will be consumed).
func NewWithArgs(blockSize uint64) Map {
	if blockSize <= 0 {
		blockSize = defaultBlockSize
	}
	blockSize = fixBlockSize(blockSize)
	result := &openAddressGrowingMap{initialSize: blockSize, threadSafety: threadSafe}
	if err := result.growTo(blockSize); err != nil {
		panic(err)
	}
	result.SetForbidGrowing(forbidGrowing)
	return result
}

func newWithArgsIface(blockSize uint64) iMap {
	return NewWithArgs(blockSize)
}

func (m *openAddressGrowingMap) SetThreadSafety(threadSafety bool) {
	m.threadSafety = threadSafety
}

func (m *openAddressGrowingMap) IsForbiddenToGrow() bool {
	return atomic.LoadInt32(&m.forbidGrowing) != 0
}

func (m *openAddressGrowingMap) SetForbidGrowing(forbidGrowing bool) {
	if forbidGrowing {
		atomic.StoreInt32(&m.forbidGrowing, 1)
	} else {
		if atomic.LoadInt32(&m.forbidGrowing) != 0 {
			panic(`Not supported, yet: you cannot reenable growing`)
		}
		atomic.StoreInt32(&m.forbidGrowing, 0)
	}
}

/*func (m *openAddressGrowingMap) increaseConcurrency() {
	if !m.threadSafety {
		return
	}
	if m.IsForbiddenToGrow() {
		return
	}
}
func (m *openAddressGrowingMap) decreaseConcurrency() {
	if !m.threadSafety {
		return
	}
	if m.IsForbiddenToGrow() {
		return
	}
}*/

func (m *openAddressGrowingMap) lock() {
	if !m.threadSafety {
		return
	}
	m.locker.Lock()
}
func (m *openAddressGrowingMap) unlock() {
	if !m.threadSafety {
		return
	}
	m.locker.Unlock()
}

type openAddressGrowingMap struct {
	*storage

	initialSize      uint64
	busySlots        int64
	writeConcurrency int32
	threadSafety     bool
	forbidGrowing    int32
	isGrowing        int32
	locker           spinlock.Locker
}

func (m *openAddressGrowingMap) waitUntilNoWrite() {
	for atomic.LoadInt32(&m.writeConcurrency) != 0 {
		time.Sleep(lockSleepInterval)
	}
}

func (m *openAddressGrowingMap) isEnoughFreeSpace() bool {
	return float64(m.BusySlots()+uint64(atomic.LoadInt32(&m.writeConcurrency)))/float64(len(m.items)) < growAtFullness
}
func (m *openAddressGrowingMap) concedeToGrowing() {
	for atomic.LoadInt32(&m.isGrowing) != 0 {
		time.Sleep(lockSleepInterval)
	}
}
func (m *openAddressGrowingMap) SetBytesByBytes(key []byte, value []byte) error {
	return m.set(func() (uint64, uint8, bool) {
		return hasher.PreHashBytes(key)
	}, func(slot *mapSlot) bool {
		return hasher.IsEqualKey(slot.key, key)
	}, func(slot *mapSlot) {
		slot.key = key
	}, func(slot *mapSlot) {
		slot.bytesValue = value
	})
}
func (m *openAddressGrowingMap) Set(key Key, value interface{}) error {
	return m.set(func() (uint64, uint8, bool) {
		return hasher.PreHash(key)
	}, func(slot *mapSlot) bool {
		return hasher.IsEqualKey(slot.key, key)
	}, func(slot *mapSlot) {
		slot.key = key
	}, func(slot *mapSlot) {
		slot.value = value
	})
}
func (m *openAddressGrowingMap) Swap(key Key, value interface{}) (oldValue interface{}, err error) {
	err = m.set(func() (uint64, uint8, bool) {
		return hasher.PreHash(key)
	}, func(slot *mapSlot) bool {
		return hasher.IsEqualKey(slot.key, key)
	}, func(slot *mapSlot) {
		slot.key = key
	}, func(slot *mapSlot) {
		oldValue = slot.value
		slot.value = value
	})
	return
}

func (m *openAddressGrowingMap) set(getPreHash func() (uint64, uint8, bool), compareKey func(*mapSlot) bool, setKey func(*mapSlot), setValue func(*mapSlot)) error {
	/*if m.currentSize == len(m.storage) {
		return NoSpaceLeft
	}*/
	if m.threadSafety {
		m.concedeToGrowing()
		if !m.isEnoughFreeSpace() {
			if err := m.growTo(m.size() << 1); err != nil {
				return err
			}
		}
		atomic.AddInt32(&m.writeConcurrency, 1)
		//m.increaseConcurrency()
	}

	preHashValue, typeID, preHashValueIsFull := getPreHash()
	hashValue := hasher.CompleteHash(preHashValue, typeID)
	idxValue := m.getIdx(hashValue)
	if !preHashValueIsFull {
		typeID = 0
	}

	var slot *mapSlot
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slot = &m.items[idxValue].mapSlot
		if slot.isSet.CompareAndSwap(isSet_notSet, isSet_setting) {
			break
		} else {
			if slot.isSet.CompareAndSwap(isSet_removed, isSet_setting) {
				break
			}
		}
		if m.threadSafety {
			if !slot.setIsUpdating() {
				if slot.isSet.CompareAndSwap(isSet_removed, isSet_setting) {
					break
				} else {
					continue // try again
				}
			}
		}
		if slot.hashValue == hashValue {
			var isEqualKey bool
			if typeID != 0 || slot.fastKeyType != 0 {
				isEqualKey = slot.fastKey == preHashValue && slot.fastKeyType == typeID
			} else {
				isEqualKey = compareKey(slot)
			}

			if isEqualKey {
				if m.threadSafety {
					slot.waitForReadersOut()
				}
				setValue(slot)
				if m.threadSafety {
					slot.isSet.Store(isSet_set)
					atomic.AddInt32(&m.writeConcurrency, -1)
					//m.decreaseConcurrency()
				}
				return nil
			}
		}
		slot.isSet.Store(isSet_set)
		slid++
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
		if slid > m.size() {
			panic(fmt.Errorf("%v %v %v %v", slid, m.size(), m.BusySlots(), m.isGrowing))
		}
	}

	slot.hashValue = hashValue
	if preHashValueIsFull {
		slot.fastKey, slot.fastKeyType = preHashValue, typeID
	}
	setKey(slot)
	setValue(slot)
	slot.slid = slid
	atomic.AddInt64(&m.busySlots, 1)
	slot.isSet.Store(isSet_set)

	if m.threadSafety {
		atomic.AddInt32(&m.writeConcurrency, -1)
		//m.decreaseConcurrency()
	}
	if !m.isEnoughFreeSpace() {
		if err := m.growTo(m.size() << 1); err != nil {
			return nil
		}
	}
	return nil
}

func copySlot(newSlot, oldSlot *mapSlot) { // is sligtly faster than "*newSlot = *oldSlot"
	newSlot.isSet = oldSlot.isSet
	newSlot.hashValue = oldSlot.hashValue
	newSlot.key = oldSlot.key
	newSlot.fastKey, newSlot.fastKeyType = oldSlot.fastKey, oldSlot.fastKeyType
	newSlot.value = oldSlot.value
}

func (m *openAddressGrowingMap) growTo(newSize uint64) error {
	if m.IsForbiddenToGrow() {
		return ForbiddenToGrow
	}

	if newSize > maximalSize {
		return NoSpaceLeft
	}

	if m.size() >= newSize {
		return nil
	}

	if m.threadSafety {
		if !atomic.CompareAndSwapInt32(&m.isGrowing, 0, 1) {
			return AlreadyGrowing
		}
		defer atomic.StoreInt32(&m.isGrowing, 0)

		m.lock()
		defer m.unlock()
		m.waitUntilNoWrite()
	}

	if m.size() >= newSize {
		return nil
	}

	newStorage := newStorage(newSize)
	newStorage.copyOldItemsAfterGrowing(m.storage)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.storage)), (unsafe.Pointer)(newStorage))
	return nil
}

func (m *openAddressGrowingMap) GetByUint64(key uint64) (interface{}, error) {
	if m.BusySlots() == 0 {
		return nil, NotFound
	}
	//m.increaseConcurrency()

	preHashValue, typeID, preHashValueIsFull := hasher.PreHashUint64(key)
	hashValue := hasher.CompleteHash(preHashValue, typeID)
	var fastKey uint64
	var fastKeyType uint8
	if preHashValueIsFull {
		fastKey, fastKeyType = preHashValue, typeID
	}
	return m.getByHashValue(fastKey, fastKeyType, hashValue, func(slot *mapSlot) bool {
		slotKey, ok := slot.key.(uint64)
		if !ok {
			return false
		}
		if slotKey != key {
			return false
		}
		return true
	})
}

func (m *openAddressGrowingMap) GetByBytes(key []byte) (interface{}, error) {
	if m.BusySlots() == 0 {
		return nil, NotFound
	}
	//m.increaseConcurrency()

	preHashValue, typeID, preHashValueIsFull := hasher.PreHashBytes(key)
	hashValue := hasher.CompleteHash(preHashValue, typeID)
	var fastKey uint64
	var fastKeyType uint8
	if preHashValueIsFull {
		fastKey, fastKeyType = preHashValue, typeID
	}
	return m.getByHashValue(fastKey, fastKeyType, hashValue, func(slot *mapSlot) bool {
		slotKey, ok := slot.key.([]byte)
		if !ok {
			return false
		}
		if len(slotKey) != len(key) {
			return false
		}
		l := len(key)
		for i := 0; i < l; i++ {
			if slotKey[i] != key[i] {
				return false
			}
		}
		return true
	})
}

func (m *openAddressGrowingMap) Get(key Key) (interface{}, error) {
	if m.BusySlots() == 0 {
		return nil, NotFound
	}
	//m.increaseConcurrency()

	preHashValue, typeID, preHashValueIsFull := hasher.PreHash(key)
	hashValue := hasher.CompleteHash(preHashValue, typeID)
	var fastKey uint64
	var fastKeyType uint8
	if preHashValueIsFull {
		fastKey, fastKeyType = preHashValue, typeID
	}
	return m.getByHashValue(fastKey, fastKeyType, hashValue, func(slot *mapSlot) bool {
		return hasher.IsEqualKey(slot.key, key)
	})
}

func (m *openAddressGrowingMap) getByHashValue(fastKey uint64, fastKeyType uint8, hashValue uint64, isRightSlotFn func(*mapSlot) bool) (interface{}, error) {
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.items[idxValue].mapSlot
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
		var isSetStatus isSet
		if m.threadSafety {
			isSetStatus = slot.increaseReaders()
		} else {
			isSetStatus = slot.IsSet()
		}
		if isSetStatus == isSet_notSet {
			break
		}
		if isSetStatus == isSet_removed {
			slot.decreaseReaders()
			continue
		}

		if slot.hashValue != hashValue {
			slot.decreaseReaders()
			continue
		}
		var isRightSlot bool
		if slot.fastKeyType != 0 || fastKeyType != 0 {
			isRightSlot = slot.fastKey == fastKey && slot.fastKeyType == fastKeyType
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

// loopy slid handler on free'ing a slot
func (m *openAddressGrowingMap) setEmptySlot(idxValue uint64, slot *mapSlot) {
	m.lock()

	// searching for a replacement to the slot (if somebody slid forward)
	slid := uint64(0)
	realRemoveIdxValue := idxValue
	freeIdxValue := idxValue
	freeSlot := slot
	for {
		slid++
		realRemoveIdxValue++
		if realRemoveIdxValue >= m.size() {
			realRemoveIdxValue = 0
		}
		realRemoveSlot := &m.items[realRemoveIdxValue].mapSlot
		if realRemoveSlot.isSet == isSet_notSet {
			break
		}
		if realRemoveSlot.slid < slid {
			continue
		}

		// searching for the last slot to move
		previousRealRemoveIdxValue := realRemoveIdxValue
		previousRealRemoveSlot := realRemoveSlot
		for {
			slid++
			realRemoveIdxValue++
			if realRemoveIdxValue >= m.size() {
				realRemoveIdxValue = 0
			}
			realRemoveSlot := &m.items[realRemoveIdxValue].mapSlot
			if realRemoveSlot.isSet == isSet_notSet {
				break
			}
			if realRemoveSlot.slid < slid {
				continue
			}
			previousRealRemoveIdxValue = realRemoveIdxValue
			previousRealRemoveSlot = realRemoveSlot
		}
		realRemoveIdxValue = previousRealRemoveIdxValue
		realRemoveSlot = previousRealRemoveSlot

		*freeSlot = *realRemoveSlot
		freeSlot.slid -= realRemoveIdxValue - freeIdxValue

		freeSlot = realRemoveSlot
		freeIdxValue = realRemoveIdxValue
		slid = 0
	}

	freeSlot.value = nil
	freeSlot.isSet = isSet_notSet
	atomic.AddInt64(&m.busySlots, -1)
	m.unlock()
}

type ConditionFunc func(value interface{}) bool

func (m *openAddressGrowingMap) unset(key Key, conditionFunc ConditionFunc) (*mapSlot, uint64) {
	preHashValue, typeID, preHashValueIsFull := hasher.PreHash(key)
	hashValue := hasher.CompleteHash(preHashValue, typeID)
	idxValue := m.getIdx(hashValue)
	if !preHashValueIsFull {
		typeID = 0
	}

	for {
		slot := &m.items[idxValue].mapSlot
		curIdxValue := idxValue
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
		switch slot.IsSet() {
		case isSet_notSet:
			return nil, 0
		case isSet_removed:
			continue
		}
		if m.threadSafety {
			if !slot.setIsUpdating() {
				continue
			}
		}
		if slot.hashValue != hashValue {
			slot.isSet.Store(isSet_set)
			continue
		}

		var isEqualKey bool
		if slot.fastKeyType != 0 || typeID != 0 {
			isEqualKey = slot.fastKey == preHashValue && slot.fastKeyType == typeID
		} else {
			isEqualKey = hasher.IsEqualKey(slot.key, key)
		}
		if !isEqualKey {
			slot.isSet.Store(isSet_set)
			continue
		}
		if conditionFunc != nil {
			var value interface{}
			if slot.bytesValue != nil {
				value = slot.bytesValue
			} else {
				value = slot.value
			}
			if !conditionFunc(value) {
				slot.isSet.Store(isSet_set)
				return nil, curIdxValue
			}
		}

		slot.isSet.Store(isSet_set)
		return slot, curIdxValue
	}
	return nil, math.MaxUint64
}
func (m *openAddressGrowingMap) Unset(key Key) error {
	return m.UnsetIf(key, nil)
}
func (m *openAddressGrowingMap) UnsetIf(key Key, conditionFunc ConditionFunc) error {
	if m.BusySlots() == 0 {
		return NotFound
	}
	//m.increaseConcurrency()
	atomic.AddInt32(&m.writeConcurrency, 1)
	//slot, idx := m.unset(key)
	slot, idx := m.unset(key, conditionFunc)
	atomic.AddInt32(&m.writeConcurrency, -1)
	//m.decreaseConcurrency()
	if slot == nil {
		if idx == math.MaxUint64 {
			return NotFound
		} else {
			return ConditionFailed
		}
	}
	//if m.IsForbiddenToGrow() {
	slot.value = nil
	slot.bytesValue = nil
	slot.isSet.Store(isSet_removed)
	atomic.AddInt64(&m.busySlots, -1)
	//} else {
	//	m.setEmptySlot(idx, slot)
	//}
	return nil
}

func (m *openAddressGrowingMap) Len() int {
	return int(atomic.LoadInt64(&m.busySlots))
}
func (m *openAddressGrowingMap) BusySlots() uint64 {
	return uint64(atomic.LoadInt64(&m.busySlots))
}

/*func (m *openAddressGrowingMap) Reset() {
	m.lock()
	*m = openAddressGrowingMap{initialSize: m.initialSize, concurrency: -1, threadSafety: threadSafe}
	m.growTo(m.initialSize)
}*/

func (m *openAddressGrowingMap) Hash(key Key) uint64 {
	return hasher.Hash(key)
}

func (m *openAddressGrowingMap) CheckConsistency() error {
	m.lock()
	count := 0
	for i := uint64(0); i < m.size(); i++ {
		slot := m.items[i].mapSlot
		if slot.isSet != isSet_set {
			continue
		}
		count++
	}

	if count != m.Len() {
		return fmt.Errorf("count != m.Len(): %v %v", count, m.Len())
	}
	m.unlock()

	for i := uint64(0); i < m.size(); i++ {
		slot := m.items[i].mapSlot
		if slot.IsSet() != isSet_set {
			continue
		}

		foundValue, err := m.Get(slot.key)
		if foundValue != slot.value || err != nil {
			hashValue := hasher.Hash(slot.key)
			expectedIdxValue := m.getIdx(hashValue)
			return fmt.Errorf("m.Get(slot.key) != slot.value: %v(%v) %v; i:%v key:%v fastkey:%v,%v expectedIdx:%v", foundValue, err, slot.value, i, slot.key, slot.fastKey, slot.fastKeyType, expectedIdxValue)
		}
	}
	return nil
}

func (m *openAddressGrowingMap) HasKey(key Key) bool {
	hashValue := hasher.Hash(key)
	idxValue := m.getIdx(hashValue)

	return m.items[idxValue].mapSlot.IsSet() == isSet_set
}

// Keys() returns a slice that contains all keys.
// If you're using Keys() in a concurrent way then keep in mind:
// Keys() scans internal storage of the map while it could be changed
// (it doesn't lock the access to the map), so you can get a mix of
// different map states from different time moments as the result
func (m *openAddressGrowingMap) Keys() []interface{} {
	r := make([]interface{}, 0, m.BusySlots())

	for idxValue := uint64(0); idxValue < m.size(); idxValue++ {
		slot := &m.items[idxValue].mapSlot
		if m.threadSafety {
			switch slot.increaseReaders() {
			case isSet_notSet, isSet_removed:
				slot.decreaseReaders()
				continue
			}
		} else {
			if slot.IsSet() != isSet_set {
				continue
			}
		}
		r = append(r, slot.key)
		if m.threadSafety {
			slot.decreaseReaders()
		}
	}

	return r
}

// ToSTDMap converts to a standart map `map[Key]interface{}`.
// If you're using ToSTDMap() in a concurrent way then keep in mind:
// ToSTDMap() scans internal storage of the map while it could be changed
// (it doesn't lock the access to the map), so you can get a mix of
// different map states from different time moments as the result
func (m *openAddressGrowingMap) ToSTDMap() map[Key]interface{} {
	r := map[Key]interface{}{}
	if m.BusySlots() == 0 {
		return r
	}
	//m.increaseConcurrency()

	for idxValue := uint64(0); idxValue < m.size(); idxValue++ {
		slot := &m.items[idxValue].mapSlot
		if m.threadSafety {
			switch slot.increaseReaders() {
			case isSet_notSet, isSet_removed:
				continue
			}
		} else {
			if slot.IsSet() != isSet_set {
				continue
			}
		}
		switch key := slot.key.(type) {
		case []byte:
			r[string(key)] = slot.value
		default:
			r[slot.key] = slot.value
		}
		if m.threadSafety {
			slot.decreaseReaders()
		}
	}

	//m.decreaseConcurrency()
	return r
}

func (m *openAddressGrowingMap) FromSTDMap(stdMap map[Key]interface{}) {
	expectedSize := uint64(float64(len(stdMap))/growAtFullness) + 1
	if expectedSize > m.initialSize {
		if err := m.growTo(powerOfTwoGE(expectedSize)); err != nil {
			panic(err)
		}
	}

	for k, v := range stdMap {
		m.Set(k, v)
	}
}
