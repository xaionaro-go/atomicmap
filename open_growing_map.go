//go:generate benchmarkCodeGen

package atomicmap

import (
	"fmt"
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
	return NewWithArgs(0, nil)
}

// blockSize should be a power of 2 and should be greater than the maximal amount of elements you're planning to store. Keep in mind: the higher blockSize you'll set the longer initialization will be (and more memory will be consumed).
func NewWithArgs(blockSize uint64, customHasher hasher.Hasher) Map {
	if blockSize <= 0 {
		blockSize = defaultBlockSize
	}
	if customHasher == nil {
		customHasher = hasher.New()
	}
	blockSize = fixBlockSize(blockSize)
	result := &openAddressGrowingMap{initialSize: blockSize, hasher: customHasher, threadSafety: threadSafe}
	if err := result.growTo(blockSize); err != nil {
		panic(err)
	}
	result.SetForbidGrowing(forbidGrowing)
	return result
}

func (m *openAddressGrowingMap) SetThreadSafety(threadSafety bool) {
	m.threadSafety = threadSafety
	m.storage.threadSafety = threadSafety
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
	hasher           hasher.Hasher
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
		return m.hasher.PreHashBytes(key)
	}, func(slot *storageItem) bool {
		return hasher.IsEqualKey(slot.key, key)
	}, func(slot *storageItem) {
		slot.key = key
	}, func(slot *storageItem) {
		slot.bytesValue = value
		slot.value = nil
	})
}
func (m *openAddressGrowingMap) Set(key Key, value interface{}) error {
	return m.set(func() (uint64, uint8, bool) {
		return m.hasher.PreHash(key)
	}, func(slot *storageItem) bool {
		return hasher.IsEqualKey(slot.key, key)
	}, func(slot *storageItem) {
		slot.key = key
	}, func(slot *storageItem) {
		slot.bytesValue = nil
		slot.value = value
	})
}

func (m *openAddressGrowingMap) set(getPreHash func() (uint64, uint8, bool), compareKey func(*storageItem) bool, setKey func(*storageItem), setValue func(*storageItem)) error {
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
	hashValue := m.hasher.CompleteHash(preHashValue, typeID)
	idxValue := m.getIdx(hashValue)

	var slot *storageItem
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slot = m.getItem(idxValue)
		if slot.IsSetCompareAndSwap(isSet_notSet, isSet_setting) {
			break
		} else {
			if slot.IsSetCompareAndSwap(isSet_removed, isSet_setting) {
				break
			}
		}
		if m.threadSafety {
			if !slot.setIsUpdating() {
				if slot.IsSetCompareAndSwap(isSet_removed, isSet_setting) {
					break
				} else {
					continue // try again
				}
			}
		}
		if slot.hashValue == hashValue {
			var isEqualKey bool
			if preHashValueIsFull || slot.fastKeyType != 0 {
				isEqualKey = slot.fastKey == preHashValue && slot.fastKeyType == typeID && preHashValueIsFull
			} else {
				isEqualKey = compareKey(slot)
			}

			if isEqualKey {
				if m.threadSafety {
					slot.waitForReadersOut()
				}
				setValue(slot)
				if m.threadSafety {
					slot.IsSetStore(isSet_set)
					atomic.AddInt32(&m.writeConcurrency, -1)
					//m.decreaseConcurrency()
				}
				return nil
			}
		}
		slot.IsSetStore(isSet_set)
		slid++
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
		if slid > m.size() { // to break an infinite loop
			panic(fmt.Errorf("%v %v %v %v", slid, m.size(), m.BusySlots(), m.isGrowing))
		}
	}

	slot.hashValue = hashValue
	if preHashValueIsFull {
		slot.fastKey, slot.fastKeyType = preHashValue, typeID
	} else {
		setKey(slot)
	}
	setValue(slot)
	slot.slid = slid
	atomic.AddInt64(&m.busySlots, 1)
	slot.IsSetStore(isSet_set)

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

func copySlot(newSlot, oldSlot *storageItem) { // is sligtly faster than "*newSlot = *oldSlot"
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

	newStorage := newStorage(newSize, m.hasher, m.threadSafety)
	newStorage.copyOldItemsAfterGrowing(m.storage)
	atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&m.storage)), (unsafe.Pointer)(newStorage))
	return nil
}

func (m *openAddressGrowingMap) GetByUint64(key uint64) (interface{}, error) {
	if m.BusySlots() == 0 {
		return nil, NotFound
	}
	//m.increaseConcurrency()

	preHashValue, typeID, preHashValueIsFull := m.hasher.PreHashUint64(key)
	return m.getByHashValue(preHashValue, typeID, preHashValueIsFull, func(slot *storageItem) bool {
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

	preHashValue, typeID, preHashValueIsFull := m.hasher.PreHashBytes(key)
	return m.getByHashValue(preHashValue, typeID, preHashValueIsFull, func(slot *storageItem) bool {
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

	preHashValue, typeID, preHashValueIsFull := m.hasher.PreHash(key)
	return m.getByHashValue(preHashValue, typeID, preHashValueIsFull, func(slot *storageItem) bool {
		return hasher.IsEqualKey(slot.key, key)
	})
}

// loopy slid handler on free'ing a slot
func (m *openAddressGrowingMap) setEmptySlot(idxValue uint64, slot *storageItem) {
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
		realRemoveSlot := m.getItem(realRemoveIdxValue)
		if isSet(realRemoveSlot.isSet) == isSet_notSet {
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
			realRemoveSlot := m.getItem(realRemoveIdxValue)
			if isSet(realRemoveSlot.isSet) == isSet_notSet {
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

func (m *openAddressGrowingMap) unset(key Key) (*storageItem, uint64) {
	preHashValue, typeID, preHashValueIsFull := m.hasher.PreHash(key)
	hashValue := m.hasher.CompleteHash(preHashValue, typeID)
	idxValue := m.getIdx(hashValue)
	if !preHashValueIsFull {
		typeID = 0
	}

	for {
		slot := m.getItem(idxValue)
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
			slot.IsSetStore(isSet_set)
			continue
		}

		var isEqualKey bool
		if slot.fastKeyType != 0 || typeID != 0 {
			isEqualKey = slot.fastKey == preHashValue && slot.fastKeyType == typeID
		} else {
			isEqualKey = hasher.IsEqualKey(slot.key, key)
		}
		if !isEqualKey {
			slot.IsSetStore(isSet_set)
			continue
		}

		slot.IsSetStore(isSet_set)
		return slot, curIdxValue
	}
	return nil, 0
}
func (m *openAddressGrowingMap) Unset(key Key) error {
	if m.BusySlots() == 0 {
		return NotFound
	}
	//m.increaseConcurrency()
	atomic.AddInt32(&m.writeConcurrency, 1)
	//slot, idx := m.unset(key)
	slot, _ := m.unset(key)
	atomic.AddInt32(&m.writeConcurrency, -1)
	//m.decreaseConcurrency()
	if slot == nil {
		return NotFound
	}
	//if m.IsForbiddenToGrow() {
	slot.IsSetStore(isSet_removed)
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
	*m = openAddressGrowingMap{initialSize: m.initialSize, hasher: m.hasher, concurrency: -1, threadSafety: threadSafe}
	m.growTo(m.initialSize)
}*/

func (m *openAddressGrowingMap) Hash(key Key) uint64 {
	return m.hasher.Hash(key)
}

func (m *openAddressGrowingMap) CheckConsistency() error {
	m.lock()
	count := 0
	for i := uint64(0); i < m.size(); i++ {
		slot := m.getItem(i)
		if isSet(slot.isSet) != isSet_set {
			continue
		}
		count++
	}

	if count != m.Len() {
		return fmt.Errorf("count != m.Len(): %v %v", count, m.Len())
	}
	m.unlock()

	for i := uint64(0); i < m.size(); i++ {
		slot := m.getItem(i)
		if slot.IsSet() != isSet_set {
			continue
		}

		var foundValue interface{}
		var err error
		if slot.fastKeyType == 0 {
			foundValue, err = m.Get(slot.key)
		} else {
			foundValue, err = m.getByHashValue(slot.fastKey, slot.fastKeyType, true, func(slot *storageItem) bool {
				return false
			})
		}
		if foundValue != slot.value || err != nil {
			hashValue := m.hasher.Hash(slot.key)
			expectedIdxValue := m.getIdx(hashValue)
			return fmt.Errorf("m.Get(slot.key) != slot.value: %v(%v) %v; i:%v key:%v fastkey:%v,%v expectedIdx:%v", foundValue, err, slot.value, i, slot.key, slot.fastKey, slot.fastKeyType, expectedIdxValue)
		}
	}
	return nil
}

func (m *openAddressGrowingMap) HasKey(key Key) bool {
	hashValue := m.hasher.Hash(key)
	idxValue := m.getIdx(hashValue)

	return m.getItem(idxValue).IsSet() == isSet_set
}

// Keys() returns a slice that contains all keys.
// If you're using Keys() in a concurrent way then keep in mind:
// Keys() scans internal storage of the map while it could be changed
// (it doesn't lock the access to the map), so you can get a mix of
// different map states from different time moments as the result
func (m *openAddressGrowingMap) Keys() []interface{} {
	r := make([]interface{}, 0, m.BusySlots())

	for idxValue := uint64(0); idxValue < m.size(); idxValue++ {
		slot := m.getItem(idxValue)
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
		slot := m.getItem(idxValue)
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
		var keyI interface{}
		if slot.fastKeyType != 0 {
			keyI = m.hasher.PreHashToKey(slot.fastKey, slot.fastKeyType)
		} else {
			keyI = slot.key
		}
		switch key := keyI.(type) {
		case []byte:
			r[string(key)] = slot.value
		default:
			r[key] = slot.value
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
