//go:generate benchmarkCodeGen

package openAddressGrowingMap

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"git.dx.center/trafficstars/testJob0/internal/errors"
	"git.dx.center/trafficstars/testJob0/internal/routines"
	I "git.dx.center/trafficstars/testJob0/task/interfaces"
)

const (
	startGrowAtFullness       = 0.73
	waitForGrowAtFullness     = 0.85
	maximalSize               = 1 << 32
	backgroundGrowOfBigSlices = false
	smallSliceSize            = 1 << 16
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

	filledIdx filledIdx
	mapValue  mapValue
}

type mapValue struct {
	isSet        bool
	hashValue    int
	filledIdxIdx uint64
	slid         uint64 // how much items were already busy so we were have to go forward
	key          I.Key
	value        interface{}
}

type filledIdx struct {
	idxValue   uint64
	whenToMove int
}

type openAddressGrowingMap struct {
	initialSize        uint64
	storage            []storageItem
	newStorage         []storageItem
	hashFunc           func(blockSize int, key I.Key) int
	busySlots          uint64
	currentGrowingStep int
	mutex              *sync.Mutex
	growMutex          *sync.Mutex
	growConcurrency    int32
}

type storageDumpItem struct {
	Key   I.Key
	Value interface{}
}

type storageDump struct {
	StorageDumpItems []storageDumpItem
}

func getIdxHashMask(size uint64) uint64 { // this function requires size to be a power of 2
	return size - 1 // example 01000000 -> 00111111
}

func (m *openAddressGrowingMap) lock() {
	m.mutex.Lock()
}

func (m *openAddressGrowingMap) unlock() {
	m.mutex.Unlock()
}

func (m openAddressGrowingMap) size() uint64 {
	return uint64(len(m.storage))
}

func (m openAddressGrowingMap) getIdx(hashValue int) uint64 {
	return uint64(hashValue) & getIdxHashMask(m.size())
}

func (m openAddressGrowingMap) getWhenToMove(currentIdxValue uint64, hashValue int) int {
	nextStep := m.currentGrowingStep
	nextIdxValue := currentIdxValue
	nextSize := m.size()
	for nextIdxValue == currentIdxValue {
		nextStep++
		nextSize = nextSize << 1
		if nextSize == 0 {
			return -1
		}
		nextIdxValue = uint64(hashValue) & getIdxHashMask(nextSize)
	}

	return nextStep
}

func (m *openAddressGrowingMap) Set(key I.Key, value interface{}) error {
	m.lock()
	defer m.unlock()
	/*if m.currentSize == len(m.storage) {
		return errors.NoSpaceLeft
	}*/

	hashValue := m.Hash(key)
	realIdxValue := m.getIdx(hashValue)
	idxValue := realIdxValue

	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		value := &m.storage[idxValue].mapValue
		if !value.isSet {
			break
		}
		if value.hashValue == hashValue {
			if routines.IsEqualKey(value.key, key) {
				value.value = value
				return nil
			}
		}
		slid++
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
	}

	whenToMove := m.getWhenToMove(idxValue, hashValue)

	item := &m.storage[idxValue].mapValue
	item.isSet = true
	item.hashValue = hashValue
	item.key = key
	item.value = value
	item.filledIdxIdx = m.busySlots
	item.slid = slid

	filledIdx := &m.storage[m.busySlots].filledIdx
	filledIdx.idxValue = idxValue
	filledIdx.whenToMove = whenToMove
	m.busySlots++

	if backgroundGrowOfBigSlices && len(m.storage) > smallSliceSize {

		if float64(m.busySlots)/float64(len(m.storage)) >= startGrowAtFullness {
			err := m.startGrow()
			if err != nil {
				return err
			}
		}

		if float64(m.busySlots)/float64(len(m.storage)) >= waitForGrowAtFullness {
			m.finishGrow()
		}
	} else {
		if float64(m.busySlots)/float64(len(m.storage)) >= waitForGrowAtFullness {
			m.growTo(m.size() << 1)
		}
	}

	return nil
}

func copySlot(newSlot, oldSlot *mapValue) { // is sligtly faster than "*newSlot = *oldSlot"
	newSlot.isSet = oldSlot.isSet
	newSlot.hashValue = oldSlot.hashValue
	newSlot.key = oldSlot.key
	newSlot.value = oldSlot.value
	newSlot.filledIdxIdx = oldSlot.filledIdxIdx
	newSlot.slid = oldSlot.slid
}
func copyFilledIdx(newFilledIdx, oldFilledIdx *filledIdx) { // is sligtly faster than "*newFilledIdx = *oldFilledIdx"
	newFilledIdx.idxValue = oldFilledIdx.idxValue
	newFilledIdx.whenToMove = oldFilledIdx.whenToMove
}

func (m *openAddressGrowingMap) updateIdx(oldIdxValue uint64) {
	oldSlot := &m.storage[oldIdxValue].mapValue
	filledIdxIdx := oldSlot.filledIdxIdx
	hashValue := oldSlot.hashValue
	newIdxValue := m.getIdx(hashValue)
	whenToMove := m.getWhenToMove(newIdxValue, hashValue)

	if oldIdxValue == newIdxValue { // TODO: comment-out this after tests
		panic(fmt.Errorf("This shouldn't happened! %v %v", oldIdxValue, m.size()))
	}

	var newSlot *mapValue
	for { // Going forward through the storage while a collision (to find a free slots)
		newSlot = &m.storage[newIdxValue].mapValue
		if !newSlot.isSet {
			break
		}
		newIdxValue++
		if newIdxValue >= m.size() {
			newIdxValue = 0
		}
	}

	copySlot(newSlot, oldSlot)
	oldSlot.isSet = false

	filledIdx := &m.storage[filledIdxIdx].filledIdx
	filledIdx.whenToMove = whenToMove
	filledIdx.idxValue = newIdxValue
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
	m.currentGrowingStep++
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
	m.currentGrowingStep++
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

func (m *openAddressGrowingMap) copyOldItemsAfterGrowing(oldStorage []storageItem) {
	for i := uint64(0); i < m.busySlots; i++ {
		filledIdx := &oldStorage[i].filledIdx
		idxValue := filledIdx.idxValue
		copySlot(&m.storage[idxValue].mapValue, &oldStorage[idxValue].mapValue)
		copyFilledIdx(&m.storage[i].filledIdx, &oldStorage[i].filledIdx)
	}

	for i := uint64(0); i < m.busySlots; i++ {
		filledIdx := &m.storage[i].filledIdx
		if filledIdx.whenToMove != m.currentGrowingStep {
			continue
		}

		m.updateIdx(filledIdx.idxValue)
	}
}

func (m *openAddressGrowingMap) Get(key I.Key) (interface{}, error) {
	m.lock()
	defer m.unlock()

	hashValue := m.Hash(key)
	idxValue := m.getIdx(hashValue)

	for {
		value := m.storage[idxValue].mapValue
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
		if !value.isSet {
			break
		}
		if value.hashValue != hashValue {
			continue
		}
		if !routines.IsEqualKey(value.key, key) {
			continue
		}
		return value.value, nil
	}

	return nil, errors.NotFound
}

func (m *openAddressGrowingMap) Unset(key I.Key) error {
	m.lock()
	defer m.unlock()

	hashValue := m.Hash(key)
	idxValue := m.getIdx(hashValue)

	for {
		value := &m.storage[idxValue].mapValue
		idxValue++
		if idxValue >= m.size() {
			idxValue = 0
		}
		if !value.isSet {
			break
		}
		if value.hashValue != hashValue {
			continue
		}
		if !routines.IsEqualKey(value.key, key) {
			continue
		}

		m.busySlots--

		// searching for a replacement to the slot (if somebody slid forward)
		slid := uint64(0)
		realRemoveIdxValue := idxValue
		for {
			slid++
			realRemoveIdxValue++
			if realRemoveIdxValue >= m.size() {
				realRemoveIdxValue = 0
			}
			realRemoveSlot := &m.storage[realRemoveIdxValue].mapValue
			if !realRemoveSlot.isSet {
				break
			}
			if realRemoveSlot.slid >= slid {
				filledIdx := &m.storage[realRemoveSlot.filledIdxIdx].filledIdx
				*filledIdx = m.storage[m.busySlots].filledIdx
				m.storage[filledIdx.idxValue].mapValue.filledIdxIdx = realRemoveSlot.filledIdxIdx

				*value = *realRemoveSlot
				realRemoveSlot.isSet = false
				return nil
			}
		}

		value.isSet = false
		filledIdx := &m.storage[value.filledIdxIdx].filledIdx
		*filledIdx = m.storage[m.busySlots].filledIdx
		m.storage[filledIdx.idxValue].mapValue.filledIdxIdx = value.filledIdxIdx
		return nil
	}

	return errors.NotFound
}
func (m openAddressGrowingMap) Count() int {
	return int(m.busySlots)
}
func (m *openAddressGrowingMap) Reset() {
	m.growLock()
	m.lock()
	*m = openAddressGrowingMap{initialSize: m.initialSize, hashFunc: m.hashFunc, mutex: &sync.Mutex{}, growMutex: &sync.Mutex{}}
	m.growTo(m.initialSize)
}
func (m *openAddressGrowingMap) DumpJson() ([]byte, error) {
	dump := storageDump{}
	dump.StorageDumpItems = make([]storageDumpItem, m.busySlots)
	for i := 0; uint64(i) < m.busySlots; i++ {
		filledIdx := &m.storage[i].filledIdx
		item := &m.storage[filledIdx.idxValue].mapValue
		dump.StorageDumpItems[i].Key = item.key
		dump.StorageDumpItems[i].Value = item.value
	}
	return json.Marshal(dump)
}

func (m openAddressGrowingMap) Hash(key I.Key) int {
	return m.hashFunc(maximalSize, key)
}

func (m *openAddressGrowingMap) CheckConsistency() error {
	m.lock()
	defer m.unlock()

	for i := uint64(0); i < m.busySlots; i++ {
		filledIdx := m.storage[i].filledIdx
		value := m.storage[filledIdx.idxValue].mapValue
		if !value.isSet {
			return fmt.Errorf("!value.isSet: %v: %v, %v", i, value, filledIdx)
		}
	}

	count := 0
	for i := uint64(0); i < m.size(); i++ {
		value := m.storage[i].mapValue
		if !value.isSet {
			continue
		}
		count++
		idxValue := m.storage[value.filledIdxIdx].filledIdx.idxValue
		if i != idxValue {
			return fmt.Errorf("i != idxValue: %v %v", i, idxValue)
		}
	}

	if count != m.Count() {
		return fmt.Errorf("count != m.Count(): %v %v", count, m.Count())
	}

	return nil
}

func (m *openAddressGrowingMap) HasCollisionWithKey(key I.Key) bool {
	m.lock()
	defer m.unlock()

	hashValue := m.Hash(key)
	idxValue := m.getIdx(hashValue)

	return m.storage[idxValue].mapValue.isSet
}
