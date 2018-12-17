//go:generate benchmarkCodeGen

package atomicmap

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/xaionaro-go/atomicmap/hash"
)

const (
	growAtFullness    = 0.85
	maximalSize       = 1 << 32
	lockSleepInterval = time.Millisecond * 10
	defaultBlockSize  = 65536
)

const (
	isSet_notSet   = 0
	isSet_set      = 1
	isSet_setting  = 2
	isSet_updating = 3
)

var (
	threadSafe = true
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
func NewWithArgs(blockSize uint64, keyHashFunc func(blockSize uint64, key Key) uint64) Map {
	if blockSize <= 0 {
		blockSize = defaultBlockSize
	}
	if keyHashFunc == nil {
		keyHashFunc = hash.KeyHashFunc
	}
	blockSize = fixBlockSize(blockSize)
	result := &openAddressGrowingMap{initialSize: blockSize, keyHashFunc: keyHashFunc, threadSafety: threadSafe}
	result.growTo(blockSize)
	return result
}

func (m *openAddressGrowingMap) SetThreadSafety(threadSafety bool) {
	m.threadSafety = threadSafety
}

func (m *openAddressGrowingMap) increaseConcurrency() {
	if !m.threadSafety {
		return
	}
	if atomic.AddInt32(&m.concurrency, 1) > 0 {
		return
	}
	atomic.AddInt32(&m.concurrency, -1)
	runtime.Gosched()
	for atomic.AddInt32(&m.concurrency, 1) <= 0 {
		atomic.AddInt32(&m.concurrency, -1)
		time.Sleep(lockSleepInterval)
	}
}
func (m *openAddressGrowingMap) decreaseConcurrency() {
	if !m.threadSafety {
		return
	}
	atomic.AddInt32(&m.concurrency, -1)
}

func (m *openAddressGrowingMap) lock() {
	if !m.threadSafety {
		return
	}
	for atomic.AddInt32(&m.concurrency, -1) != -1 {
		atomic.AddInt32(&m.concurrency, 1)
		runtime.Gosched()
	}
}
func (m *openAddressGrowingMap) unlock() {
	if !m.threadSafety {
		return
	}
	atomic.AddInt32(&m.concurrency, 1) // back to zero
}

type storageItem struct {
	// All the stuff what we need to grow. This variables are not connected
	// It's just all (unrelated) slices we need united into one to decrease
	// the number of memory allocations

	mapSlot mapSlot
	// ...other variables here...
}

type mapSlot struct {
	isSet        uint32
	readersCount int32
	hashValue    uint64
	slid         uint64 // how much items were already busy so we were have to go forward
	key          Key
	value        interface{}
}

func (slot *mapSlot) IsSet() uint32 {
	return atomic.LoadUint32(&slot.isSet)
}

func (slot *mapSlot) waitForIsSet() bool {
	switch slot.IsSet() {
	case isSet_set:
		return true
	case isSet_notSet:
		return false
	}

	runtime.Gosched()
	for {
		switch slot.IsSet() {
		case isSet_set:
			return true
		case isSet_notSet:
			return false
		}
		time.Sleep(lockSleepInterval)
	}
}

func (slot *mapSlot) setIsUpdating() {
	if atomic.CompareAndSwapUint32(&slot.isSet, isSet_set, isSet_updating) {
		return
	}

	runtime.Gosched()
	for !atomic.CompareAndSwapUint32(&slot.isSet, isSet_set, isSet_updating) {
		time.Sleep(lockSleepInterval)
	}
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

func (slot *mapSlot) increaseReaders() bool {
	if !slot.waitForIsSet() {
		return false
	}
	atomic.AddInt32(&slot.readersCount, 1)
	if !slot.waitForIsSet() {
		atomic.AddInt32(&slot.readersCount, -1)
		return false
	}
	return true
}

func (slot *mapSlot) decreaseReaders() {
	atomic.AddInt32(&slot.readersCount, -1)
}

type openAddressGrowingMap struct {
	initialSize    uint64
	storage        []storageItem
	newStorage     []storageItem
	keyHashFunc    func(blockSize uint64, key Key) uint64
	busySlots      int64
	setConcurrency int32
	concurrency    int32
	threadSafety   bool
	isGrowing      int32
}

func (m *openAddressGrowingMap) size() uint64 {
	return uint64(len(m.storage))
}

func getIdxHashMask(size uint64) uint64 { // this function requires size to be a power of 2
	return size - 1 // example 01000000 -> 00111111
}

func (m *openAddressGrowingMap) getIdx(hashValue uint64) uint64 {
	return hashValue & getIdxHashMask(m.size())
}

func (m *openAddressGrowingMap) isEnoughFreeSpace() bool {
	return float64(m.BusySlots()+uint64(atomic.LoadInt32(&m.setConcurrency)))/float64(len(m.storage)) < growAtFullness
}
func (m *openAddressGrowingMap) concedeToGrowing() {
	for atomic.LoadInt32(&m.isGrowing) != 0 {
		time.Sleep(lockSleepInterval)
	}
}
func (m *openAddressGrowingMap) Set(key Key, value interface{}) error {
	/*if m.currentSize == len(m.storage) {
		return NoSpaceLeft
	}*/
	if m.threadSafety {
		atomic.AddInt32(&m.setConcurrency, 1)
		if !m.isEnoughFreeSpace() {
			m.growTo(m.size() << 1)
		}
		m.concedeToGrowing()
		m.increaseConcurrency()
	}

	hashValue := hash.Uint64Hash(maximalSize, m.keyHashFunc(maximalSize, key))
	idxValue := m.getIdx(hashValue)

	var slot *mapSlot
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slot = &m.storage[idxValue].mapSlot
		if atomic.CompareAndSwapUint32(&slot.isSet, isSet_notSet, isSet_setting) {
			break
		}
		if m.threadSafety {
			slot.setIsUpdating()
		}
		if slot.hashValue == hashValue {
			if hash.IsEqualKey(slot.key, key) {
				if m.threadSafety {
					slot.waitForReadersOut()
				}
				slot.value = value
				if m.threadSafety {
					atomic.StoreUint32(&slot.isSet, isSet_set)
					atomic.AddInt32(&m.setConcurrency, -1)
					m.decreaseConcurrency()
				}
				return nil
			}
		}
		atomic.StoreUint32(&slot.isSet, isSet_set)
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
	slot.key = key
	slot.value = value
	slot.slid = slid
	atomic.StoreUint32(&slot.isSet, isSet_set)

	atomic.AddInt64(&m.busySlots, 1)

	if m.threadSafety {
		atomic.AddInt32(&m.setConcurrency, -1)
		m.decreaseConcurrency()
	}
	if !m.isEnoughFreeSpace() {
		m.growTo(m.size() << 1)
	}
	return nil
}

func copySlot(newSlot, oldSlot *mapSlot) { // is sligtly faster than "*newSlot = *oldSlot"
	newSlot.isSet = oldSlot.isSet
	newSlot.hashValue = oldSlot.hashValue
	newSlot.key = oldSlot.key
	newSlot.value = oldSlot.value
}

func (m *openAddressGrowingMap) growTo(newSize uint64) error {
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
	}

	if m.size() >= newSize {
		return nil
	}

	oldSize := m.size()
	oldStorage := m.storage
	m.storage = make([]storageItem, newSize)
	if oldSize == 0 {
		return nil
	}

	m.copyOldItemsAfterGrowing(oldStorage)
	return nil
}

func (m *openAddressGrowingMap) findFreeSlot(idxValue uint64) (*mapSlot, uint64, uint64) {
	var slotCandidate *mapSlot
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slotCandidate = &m.storage[idxValue].mapSlot
		if slotCandidate.isSet == isSet_notSet {
			return slotCandidate, idxValue, slid
		}
		slid++
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
	}
}

func (m *openAddressGrowingMap) copyOldItemsAfterGrowing(oldStorage []storageItem) {
	for i := 0; i < len(oldStorage); i++ {
		oldSlot := &oldStorage[i].mapSlot
		if oldSlot.isSet == isSet_notSet {
			continue
		}

		newIdxValue := m.getIdx(oldSlot.hashValue)
		newSlot, _, slid := m.findFreeSlot(newIdxValue)
		copySlot(newSlot, oldSlot)
		newSlot.slid = slid
	}
}
func (m *openAddressGrowingMap) Get(key Key) (interface{}, error) {
	if m.BusySlots() == 0 {
		return nil, NotFound
	}
	m.increaseConcurrency()

	hashValue := hash.Uint64Hash(maximalSize, m.keyHashFunc(maximalSize, key))
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.storage[idxValue].mapSlot
		if m.threadSafety {
			if !slot.increaseReaders() {
				break
			}
		} else {
			if atomic.LoadUint32(&slot.isSet) == isSet_notSet {
				break
			}
		}

		if slot.hashValue != hashValue {
			slot.decreaseReaders()
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			continue
		}
		if !hash.IsEqualKey(slot.key, key) {
			slot.decreaseReaders()
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			continue
		}

		value := slot.value
		slot.decreaseReaders()
		m.decreaseConcurrency()
		return value, nil
	}

	m.decreaseConcurrency()
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
		realRemoveSlot := &m.storage[realRemoveIdxValue].mapSlot
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
			realRemoveSlot := &m.storage[realRemoveIdxValue].mapSlot
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

	freeSlot.isSet = isSet_notSet
	atomic.AddInt64(&m.busySlots, -1)
	m.unlock()
}

func (m *openAddressGrowingMap) unset(key Key) (*mapSlot, uint64) {
	hashValue := hash.Uint64Hash(maximalSize, m.keyHashFunc(maximalSize, key))
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.storage[idxValue].mapSlot
		if atomic.LoadUint32(&slot.isSet) == isSet_notSet {
			break
		}
		if m.threadSafety {
			slot.setIsUpdating()
		}
		if slot.hashValue != hashValue {
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			atomic.StoreUint32(&slot.isSet, isSet_set)
			continue
		}
		if !hash.IsEqualKey(slot.key, key) {
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			atomic.StoreUint32(&slot.isSet, isSet_set)
			continue
		}

		atomic.StoreUint32(&slot.isSet, isSet_set)
		return slot, idxValue
	}
	return nil, 0
}
func (m *openAddressGrowingMap) LockUnset(key Key) error {
	if m.BusySlots() == 0 {
		return NotFound
	}
	m.lock()
	slot, idx := m.unset(key)
	m.unlock()
	if slot == nil {
		return NotFound
	}
	m.setEmptySlot(idx, slot)
	return nil
}
func (m *openAddressGrowingMap) Unset(key Key) error {
	if m.BusySlots() == 0 {
		return NotFound
	}
	if atomic.LoadInt32(&m.concurrency) != 0 {
		panic("Thread-safety for Unset() is not implemented, yet")
	}
	m.increaseConcurrency()
	slot, idx := m.unset(key)
	m.decreaseConcurrency()
	if slot == nil {
		return NotFound
	}
	m.setEmptySlot(idx, slot)
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
	*m = openAddressGrowingMap{initialSize: m.initialSize, keyHashFunc: m.keyHashFunc, concurrency: -1, threadSafety: threadSafe}
	m.growTo(m.initialSize)
}*/

func (m *openAddressGrowingMap) Hash(key Key) uint64 {
	return m.keyHashFunc(maximalSize, key)
}

func (m *openAddressGrowingMap) CheckConsistency() error {
	m.lock()
	count := 0
	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if slot.isSet == isSet_notSet {
			continue
		}

		count++
	}

	if count != m.Len() {
		return fmt.Errorf("count != m.Len(): %v %v", count, m.Len())
	}
	m.unlock()

	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if slot.isSet == isSet_notSet {
			continue
		}

		foundValue, err := m.Get(slot.key)
		if foundValue != slot.value || err != nil {
			hashValue := hash.Uint64Hash(maximalSize, m.keyHashFunc(maximalSize, slot.key))
			expectedIdxValue := m.getIdx(hashValue)
			return fmt.Errorf("m.Get(slot.key) != slot.value: %v(%v) %v; i:%v key:%v expectedIdx:%v", foundValue, err, slot.value, i, slot.key, expectedIdxValue)
		}
	}
	return nil
}

func (m *openAddressGrowingMap) HasCollisionWithKey(key Key) bool {
	hashValue := hash.Uint64Hash(maximalSize, m.keyHashFunc(maximalSize, key))
	idxValue := m.getIdx(hashValue)

	return m.storage[idxValue].mapSlot.isSet != isSet_notSet
}

// Keys() returns a slice that contains all keys.
// If you're using Keys() in a concurrent way then keep in mind:
// Keys() scans internal storage of the map while it could be changed
// (it doesn't lock the access to the map), so you can get a mix of
// different map states from different time moments as the result
func (m *openAddressGrowingMap) Keys() []interface{} {
	r := make([]interface{}, 0, m.BusySlots())

	for idxValue := uint64(0); idxValue < m.size(); idxValue++ {
		slot := &m.storage[idxValue].mapSlot
		if m.threadSafety {
			if !slot.increaseReaders() {
				continue
			}
		} else {
			if atomic.LoadUint32(&slot.isSet) == isSet_notSet {
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
	m.increaseConcurrency()

	for idxValue := uint64(0); idxValue < m.size(); idxValue++ {
		slot := &m.storage[idxValue].mapSlot
		if m.threadSafety {
			if !slot.increaseReaders() {
				continue
			}
		} else {
			if atomic.LoadUint32(&slot.isSet) == isSet_notSet {
				continue
			}
		}
		r[slot.key] = slot.value
		if m.threadSafety {
			slot.decreaseReaders()
		}
	}

	m.decreaseConcurrency()
	return r
}

func (m *openAddressGrowingMap) FromSTDMap(stdMap map[Key]interface{}) {
	expectedSize := uint64(float64(len(stdMap))/growAtFullness) + 1
	if expectedSize > m.initialSize {
		m.growTo(powerOfTwoGE(expectedSize))
	}

	for k, v := range stdMap {
		m.Set(k, v)
	}
}
