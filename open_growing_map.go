//go:generate benchmarkCodeGen

package openAddressGrowingMap

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"

	I "github.com/xaionaro-go/atomicmap/interfaces"
	"github.com/xaionaro-go/atomicmap/internal/errors"
	"github.com/xaionaro-go/atomicmap/internal/routines"
)

const (
	growAtFullness    = 0.85
	maximalSize       = 1 << 32
	lockSleepInterval = time.Millisecond * 10

	isSet_notSet   = 0
	isSet_set      = 1
	isSet_setting  = 2
	isSet_updating = 3
)

var (
	threadSafe = true
)

func fixBlockSize(blockSizeRaw int) (blockSize uint64) {
	// This functions fixes blockSize value to be a power of 2

	if blockSizeRaw <= 0 {
		log.Printf("Invalid block size: %v. Setting to 1024\n", blockSize)
		blockSize = 1024
	} else {
		blockSize = uint64(blockSizeRaw)
	}

	if (blockSize-1)^blockSize < blockSize {
		shiftedBlockSize := blockSize
		for (blockSize+1)^blockSize < (blockSize + 1) {
			shiftedBlockSize >>= 1
			blockSize |= shiftedBlockSize
		}
		blockSize++
		log.Printf("blockSize should be a power of 2 (1, 2, 4, 8, 16, ...). Setting to %v", blockSize)
	}

	return
}

func NewHashMap(blockSizeRaw int, fn func(blockSize int, key I.Key) int) I.HashMaper {
	blockSize := fixBlockSize(blockSizeRaw)
	result := &openAddressGrowingMap{initialSize: blockSize, hashFunc: fn, threadSafety: threadSafe}
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
	isSet     uint32
	hashValue uint64
	slid      uint64 // how much items were already busy so we were have to go forward
	key       I.Key
	value     interface{}
}

type openAddressGrowingMap struct {
	initialSize    uint64
	storage        []storageItem
	newStorage     []storageItem
	hashFunc       func(blockSize int, key I.Key) int
	busySlots      uint64
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
	return float64(m.busySlots+uint64(atomic.LoadInt32(&m.setConcurrency)))/float64(len(m.storage)) < growAtFullness
}
func (m *openAddressGrowingMap) concedeToGrowing() {
	for atomic.LoadInt32(&m.isGrowing) != 0 {
		time.Sleep(lockSleepInterval)
	}
}
func (m *openAddressGrowingMap) Set(key I.Key, value interface{}) error {
	/*if m.currentSize == len(m.storage) {
		return errors.NoSpaceLeft
	}*/
	if m.threadSafety {
		atomic.AddInt32(&m.setConcurrency, 1)
		if !m.isEnoughFreeSpace() {
			m.growTo(m.size() << 1)
		}
		m.concedeToGrowing()
		m.increaseConcurrency()
	}

	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	idxValue := m.getIdx(hashValue)

	var slot *mapSlot
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slot = &m.storage[idxValue].mapSlot
		if atomic.CompareAndSwapUint32(&slot.isSet, isSet_notSet, isSet_setting) {
			break
		}
		if m.threadSafety {
			if !atomic.CompareAndSwapUint32(&slot.isSet, isSet_set, isSet_updating) {
				runtime.Gosched()
				for !atomic.CompareAndSwapUint32(&slot.isSet, isSet_set, isSet_updating) {
					time.Sleep(lockSleepInterval)
				}
			}
		}
		if slot.hashValue == hashValue {
			if routines.IsEqualKey(slot.key, key) {
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
			panic(fmt.Errorf("%v %v %v %v", slid, m.size(), m.busySlots, m.isGrowing))
		}
	}

	slot.hashValue = hashValue
	slot.key = key
	slot.value = value
	slot.slid = slid
	atomic.StoreUint32(&slot.isSet, isSet_set)

	m.busySlots++

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
		return errors.NoSpaceLeft
	}

	if m.size() >= newSize {
		return nil
	}

	if m.threadSafety {
		if !atomic.CompareAndSwapInt32(&m.isGrowing, 0, 1) {
			return errors.AlreadyGrowing
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
func (m *openAddressGrowingMap) Get(key I.Key) (interface{}, error) {
	if m.busySlots == 0 {
		return nil, errors.NotFound
	}
	m.increaseConcurrency()

	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.storage[idxValue].mapSlot
		if atomic.LoadUint32(&slot.isSet) == isSet_notSet {
			break
		}
		if m.threadSafety {
			if atomic.LoadUint32(&slot.isSet) != isSet_set {
				runtime.Gosched()
				for atomic.LoadUint32(&slot.isSet) != isSet_set {
					time.Sleep(lockSleepInterval)
				}
			}
		}
		if slot.hashValue != hashValue {
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			continue
		}
		if !routines.IsEqualKey(slot.key, key) {
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			continue
		}
		m.decreaseConcurrency()
		return slot.value, nil
	}

	m.decreaseConcurrency()
	return nil, errors.NotFound
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
	m.busySlots--
	m.unlock()
}

func (m *openAddressGrowingMap) Unset(key I.Key) error {
	if m.busySlots == 0 {
		return errors.NotFound
	}
	if m.concurrency != 0 {
		panic("Thread-safety for Unset() is not implemented, yet")
	}
	m.increaseConcurrency()

	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.storage[idxValue].mapSlot
		if atomic.LoadUint32(&slot.isSet) == isSet_notSet {
			break
		}
		if m.threadSafety {
			if !atomic.CompareAndSwapUint32(&slot.isSet, isSet_set, isSet_updating) {
				runtime.Gosched()
				for !atomic.CompareAndSwapUint32(&slot.isSet, isSet_set, isSet_updating) {
					time.Sleep(lockSleepInterval)
				}
			}
		}
		if slot.hashValue != hashValue {
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			atomic.StoreUint32(&slot.isSet, isSet_set)
			continue
		}
		if !routines.IsEqualKey(slot.key, key) {
			idxValue++
			if idxValue >= m.size() {
				idxValue = 0
			}
			atomic.StoreUint32(&slot.isSet, isSet_set)
			continue
		}

		atomic.StoreUint32(&slot.isSet, isSet_set)
		m.decreaseConcurrency()
		m.setEmptySlot(idxValue, slot)
		return nil
	}

	m.decreaseConcurrency()
	return errors.NotFound
}
func (m *openAddressGrowingMap) Count() int {
	return int(m.busySlots)
}
func (m *openAddressGrowingMap) Reset() {
	m.lock()
	*m = openAddressGrowingMap{initialSize: m.initialSize, hashFunc: m.hashFunc, concurrency: -1, threadSafety: threadSafe}
	m.growTo(m.initialSize)
}

func (m *openAddressGrowingMap) Hash(key I.Key) int {
	return m.hashFunc(maximalSize, key)
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

	if count != m.Count() {
		return fmt.Errorf("count != m.Count(): %v %v", count, m.Count())
	}
	m.unlock()

	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if slot.isSet == isSet_notSet {
			continue
		}

		foundValue, err := m.Get(slot.key)
		if foundValue != slot.value || err != nil {
			hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, slot.key)))
			expectedIdxValue := m.getIdx(hashValue)
			return fmt.Errorf("m.Get(slot.key) != slot.value: %v(%v) %v; i:%v key:%v expectedIdx:%v", foundValue, err, slot.value, i, slot.key, expectedIdxValue)
		}
	}
	return nil
}

func (m *openAddressGrowingMap) HasCollisionWithKey(key I.Key) bool {
	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	idxValue := m.getIdx(hashValue)

	return m.storage[idxValue].mapSlot.isSet != isSet_notSet
}
