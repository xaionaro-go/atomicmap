//go:generate benchmarkCodeGen

package openAddressGrowingMap

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"git.dx.center/trafficstars/testJob0/internal/errors"
	"git.dx.center/trafficstars/testJob0/internal/routines"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

const (
	waitForGrowAtFullness     = 0.85
	maximalSize               = 1 << 32
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
	result := &openAddressGrowingMap{initialSize: blockSize, hashFunc: fn, mutex: &sync.Mutex{}, growMutex: &sync.Mutex{}}
	result.growTo(blockSize)
	return result
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
	initialSize     uint64
	storage         []storageItem
	newStorage      []storageItem
	hashFunc        func(blockSize int, key I.Key) int
	busySlots       uint64
	mutex           *sync.Mutex
	concurrency     int32
	growMutex       *sync.Mutex
	growConcurrency int32
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

func (m *openAddressGrowingMap) Set(key I.Key, value interface{}) error {
	/*if m.currentSize == len(m.storage) {
		return errors.NoSpaceLeft
	}*/

	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	realIdxValue := m.getIdx(hashValue)
	idxValue := realIdxValue

	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slot := &m.storage[idxValue].mapSlot
		if atomic.CompareAndSwapUint32(&slot.isSet, 0, 1) {
			break
		}
		if slot.hashValue == hashValue {
			if routines.IsEqualKey(slot.key, key) {
				slot.value = value
				return nil
			}
		}
		slid++
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
	}

	item := &m.storage[idxValue].mapSlot
	item.hashValue = hashValue
	item.key = key
	item.value = value
	item.slid = slid

	m.busySlots++

	if float64(m.busySlots)/float64(len(m.storage)) >= waitForGrowAtFullness {
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

func (m *openAddressGrowingMap) startGrow() error {
	newSize := m.size() << 1
	if newSize > maximalSize {
		return errors.NoSpaceLeft
	}

	for atomic.AddInt32(&m.growConcurrency, 1) != 1 {
		atomic.AddInt32(&m.growConcurrency, -1)
		return nil
	}

	m.growLock()
	go func() {
		defer m.growUnlock()
		m.newStorage = make([]storageItem, newSize)
	}()

	return nil
}

func (m *openAddressGrowingMap) finishGrow() {
	m.growLock()
	defer m.growUnlock()
	oldSize := m.size()
	if oldSize == 0 {
		return
	}
	atomic.AddInt32(&m.growConcurrency, -1)
	oldStorage := m.storage
	m.storage = m.newStorage
	m.copyOldItemsAfterGrowing(oldStorage)
}

func (m *openAddressGrowingMap) growTo(newSize uint64) error {
	if newSize > maximalSize {
		return errors.NoSpaceLeft
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

func (m *openAddressGrowingMap) growLock() {
	m.growMutex.Lock()
}
func (m *openAddressGrowingMap) growUnlock() {
	m.growMutex.Unlock()
}

func (m *openAddressGrowingMap) waitForGrow() {
	m.growLock()
	m.growUnlock()
}

func (m *openAddressGrowingMap) findFreeSlot(idxValue uint64) (*mapSlot, uint64, uint64) {
	var slotCandidate *mapSlot
	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		slotCandidate = &m.storage[idxValue].mapSlot
		if slotCandidate.isSet == 0 {
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
		if oldSlot.isSet == 0 {
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

	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.storage[idxValue].mapSlot
		if slot.isSet == 0 {
			break
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
		return slot.value, nil
	}

	return nil, errors.NotFound
}

// loopy slid handler on free'ing a slot
func (m *openAddressGrowingMap) setEmptySlot(idxValue uint64, slot *mapSlot) {

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
		if realRemoveSlot.isSet == 0 {
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
			if realRemoveSlot.isSet == 0 {
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

	freeSlot.isSet = 0
}

func (m *openAddressGrowingMap) Unset(key I.Key) error {
	if m.busySlots == 0 {
		return errors.NotFound
	}

	hashValue := routines.Uint64Hash(maximalSize, uint64(m.hashFunc(maximalSize, key)))
	idxValue := m.getIdx(hashValue)

	for {
		slot := &m.storage[idxValue].mapSlot
		if slot.isSet == 0 {
			break
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

		m.setEmptySlot(idxValue, slot)
		m.busySlots--
		return nil
	}

	return errors.NotFound
}
func (m *openAddressGrowingMap) Count() int {
	return int(m.busySlots)
}
func (m *openAddressGrowingMap) Reset() {
	m.growLock()
	*m = openAddressGrowingMap{initialSize: m.initialSize, hashFunc: m.hashFunc, mutex: &sync.Mutex{}, growMutex: &sync.Mutex{}}
	m.growTo(m.initialSize)
}

func (m *openAddressGrowingMap) Hash(key I.Key) int {
	return m.hashFunc(maximalSize, key)
}

func (m *openAddressGrowingMap) CheckConsistency() error {
	count := 0
	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if slot.isSet == 0 {
			continue
		}

		count++
	}

	if count != m.Count() {
		return fmt.Errorf("count != m.Count(): %v %v", count, m.Count())
	}

	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if slot.isSet == 0 {
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

	return m.storage[idxValue].mapSlot.isSet != 0
}
