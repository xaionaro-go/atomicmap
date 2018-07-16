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
	waitForGrowAtFullness     = 0.45
	maximalSize               = 1 << 32
	backgroundGrowOfBigSlices = false
	locks                     = false
	smallSliceSize            = 1 << 16
)

func init() {
	if !locks && backgroundGrowOfBigSlices {
		panic("!locks && backgroundGrowOfBigSlices")
	}
}

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

	filledIdx uint64
	mapSlot  mapSlot
}

type mapSlot struct {
	isSet        bool
	hashValue    int
	filledIdxIdx uint64
	slid         uint64 // how much items were already busy so we were have to go forward
	key          I.Key
	value        interface{}
}

type openAddressGrowingMap struct {
	initialSize        uint64
	storage            []storageItem
	newStorage         []storageItem
	hashFunc           func(blockSize int, key I.Key) int
	busySlots          uint64
	mutex              *sync.Mutex
	concurrency        int32
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

func (m *openAddressGrowingMap) lock() {
	if locks {
		m.mutex.Lock()
	}
}

func (m *openAddressGrowingMap) unlock() {
	if locks {
		m.mutex.Unlock()
	}
}

func (m *openAddressGrowingMap) size() uint64 {
	return uint64(len(m.storage))
}

func (m *openAddressGrowingMap) getIdx(hashValue int) uint64 {
	return routines.Uint64Hash(m.size(), uint64(hashValue))
}

func (m *openAddressGrowingMap) Set(key I.Key, value interface{}) error {
	m.lock()
	/*if m.currentSize == len(m.storage) {
		return errors.NoSpaceLeft
	}*/

	hashValue := m.hashFunc(maximalSize, key)
	realIdxValue := m.getIdx(hashValue)
	idxValue := realIdxValue

	slid := uint64(0)
	for { // Going forward through the storage while a collision (to find a free slots)
		value := &m.storage[idxValue].mapSlot
		if !value.isSet {
			break
		}
		if value.hashValue == hashValue {
			if routines.IsEqualKey(value.key, key) {
				value.value = value
				m.unlock()
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
	item.isSet = true
	item.hashValue = hashValue
	item.key = key
	item.value = value
	item.filledIdxIdx = m.busySlots
	item.slid = slid

	m.storage[m.busySlots].filledIdx = idxValue
	m.busySlots++

	if backgroundGrowOfBigSlices && len(m.storage) > smallSliceSize {

		if float64(m.busySlots)/float64(len(m.storage)) >= startGrowAtFullness {
			err := m.startGrow()
			if err != nil {
				m.unlock()
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

	m.unlock()
	return nil
}

func copySlot(newSlot, oldSlot *mapSlot) { // is sligtly faster than "*newSlot = *oldSlot"
	newSlot.isSet = oldSlot.isSet
	newSlot.hashValue = oldSlot.hashValue
	newSlot.key = oldSlot.key
	newSlot.value = oldSlot.value
}

func (m *openAddressGrowingMap) updateIdx(oldIdxValue uint64) {
	oldSlot := &m.storage[oldIdxValue].mapSlot
	filledIdxIdx := oldSlot.filledIdxIdx
	hashValue := oldSlot.hashValue
	newIdxValue := m.getIdx(hashValue)

	if oldIdxValue == newIdxValue {
		return
	}

	slid := uint64(0)
	var newSlot *mapSlot
	for { // Going forward through the storage while a collision (to find a free slots)
		newSlot = &m.storage[newIdxValue].mapSlot
		if !newSlot.isSet {
			break
		}
		newIdxValue++
		if newIdxValue >= m.size() {
			newIdxValue = 0
		}
	}

	if newSlot.isSet { // TODO: comment-out this after tests
		panic(fmt.Errorf("This shouldn't happened! %v %v %v", oldIdxValue, m.size(), newIdxValue))
	}

	copySlot(newSlot, oldSlot)
	newSlot.slid = slid

	freeFilledIdxIdx := m.setEmptySlot(oldIdxValue, oldSlot)
	m.storage[freeFilledIdxIdx].filledIdx = newIdxValue
	newSlot.filledIdxIdx = freeFilledIdxIdx
	fmt.Println("updateIdx: ", m.busySlots, filledIdxIdx, "<-:", newIdxValue)
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

func (m *openAddressGrowingMap) copyOldItemsAfterGrowing(oldStorage []storageItem) {
	for i := uint64(0); i < m.busySlots; i++ {
		idxValue := oldStorage[i].filledIdx
		oldSlot := &oldStorage[idxValue].mapSlot
		newSlot := &m.storage[idxValue].mapSlot
		copySlot(newSlot, oldSlot)
		newSlot.filledIdxIdx = oldSlot.filledIdxIdx
		newSlot.slid = oldSlot.slid
		m.storage[i].filledIdx = oldStorage[i].filledIdx
	}

	for i := uint64(0); i < m.busySlots; i++ {
		idxValue := m.storage[i].filledIdx
		if !m.storage[idxValue].mapSlot.isSet { // TODO: remove this
			panic(fmt.Errorf("This should't happened! %v %v", i, idxValue))
		}

		fmt.Println("updateIdx", m.size(), idxValue, m.storage[idxValue].mapSlot)
		m.updateIdx(idxValue)
	}
}

func (m *openAddressGrowingMap) Get(key I.Key) (interface{}, error) {
	m.lock()
	if m.busySlots == 0 {
		m.unlock()
		return nil, errors.NotFound
	}

	hashValue := m.hashFunc(maximalSize, key)
	idxValue := m.getIdx(hashValue)

	for {
		value := &m.storage[idxValue].mapSlot
		if !value.isSet {
			break
		}
		if value.hashValue != hashValue {
			idxValue++
			if idxValue >= m.size() { idxValue = 0 }
			continue
		}
		if !routines.IsEqualKey(value.key, key) {
			idxValue++
			if idxValue >= m.size() { idxValue = 0 }
			continue
		}
		m.unlock()
		return value.value, nil
	}

	m.unlock()
	return nil, errors.NotFound
}

func (m *openAddressGrowingMap) setEmptySlot(idxValue uint64, slot *mapSlot) uint64 {
	if !slot.isSet {
		panic(fmt.Errorf("This shouldn't happened: %v", idxValue))
	}

	fmt.Println("setEmptySlot", m.busySlots, idxValue)

	// searching for a replacement to the slot (if somebody slid forward)
	slid := uint64(0)
	realRemoveIdxValue := idxValue
	freeIdxValue := idxValue
	freeSlot := slot
	freeFilledIdxIdx := slot.filledIdxIdx
	for {
		slid++
		realRemoveIdxValue++
		if realRemoveIdxValue >= m.size() {
			realRemoveIdxValue = 0
		}
		realRemoveSlot := &m.storage[realRemoveIdxValue].mapSlot
		if !realRemoveSlot.isSet {
			break
		}
		if realRemoveSlot.slid < slid {
			break
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
			if !realRemoveSlot.isSet {
				break
			}
			if realRemoveSlot.slid < slid {
				break
			}
			previousRealRemoveIdxValue = realRemoveIdxValue
			previousRealRemoveSlot = realRemoveSlot
		}
		realRemoveIdxValue = previousRealRemoveIdxValue
		realRemoveSlot = previousRealRemoveSlot
		fmt.Println(">>", freeIdxValue, freeFilledIdxIdx, "<-", realRemoveIdxValue, realRemoveSlot.filledIdxIdx, m.storage[realRemoveSlot.filledIdxIdx].filledIdx, *realRemoveSlot)

		for i:=freeIdxValue; i<=realRemoveIdxValue; i++ {
			fmt.Println(">>dump", m.storage[i].mapSlot)
		}

		*freeSlot = *realRemoveSlot
		freeSlot.filledIdxIdx = freeFilledIdxIdx

		fmt.Println(">>>", freeIdxValue, freeFilledIdxIdx, *freeSlot, m.storage[freeFilledIdxIdx])

		freeFilledIdxIdx = realRemoveSlot.filledIdxIdx
		freeSlot = realRemoveSlot
		freeIdxValue = realRemoveIdxValue
		slid = 0
	}

	freeSlot.isSet = false
	fmt.Printf("freeSlot: freeIdxValue:%v freeFilledIdxIdx:%v freeSlot:%v\n", freeIdxValue, freeFilledIdxIdx, freeSlot)
	return freeFilledIdxIdx
}

func (m *openAddressGrowingMap) Unset(key I.Key) error {
	m.lock()
	if m.busySlots == 0 {
		m.unlock()
		return errors.NotFound
	}

	hashValue := m.hashFunc(maximalSize, key)
	idxValue := m.getIdx(hashValue)

	for {
		value := &m.storage[idxValue].mapSlot
		if !value.isSet {
			break
		}
		if value.hashValue != hashValue {
			idxValue++
			if idxValue >= m.size() { idxValue = 0 }
			continue
		}
		if !routines.IsEqualKey(value.key, key) {
			idxValue++
			if idxValue >= m.size() { idxValue = 0 }
			continue
		}

		freeFilledIdxIdx := m.setEmptySlot(idxValue, value)
		m.busySlots--
		filledIdx := m.storage[m.busySlots].filledIdx
		fmt.Println("at the tail:", m.storage[m.busySlots].filledIdx)
		m.storage[freeFilledIdxIdx].filledIdx = filledIdx
		fmt.Println("setEmptySlot: ", m.storage[filledIdx].mapSlot.filledIdxIdx, "<=", "<<-", m.busySlots, ":", filledIdx)
		m.storage[filledIdx].mapSlot.filledIdxIdx = freeFilledIdxIdx
		fmt.Printf("result: f.idxValue:%v check.filledIdxIdx:%v\n", filledIdx, m.storage[filledIdx].mapSlot.filledIdxIdx)
		m.unlock()
		return nil
	}

	m.unlock()
	return errors.NotFound
}
func (m *openAddressGrowingMap) Count() int {
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
		idxValue := m.storage[i].filledIdx
		item := &m.storage[idxValue].mapSlot
		dump.StorageDumpItems[i].Key = item.key
		dump.StorageDumpItems[i].Value = item.value
	}
	return json.Marshal(dump)
}

func (m *openAddressGrowingMap) Hash(key I.Key) int {
	return m.hashFunc(maximalSize, key)
}

func (m *openAddressGrowingMap) CheckConsistency() error {
	m.lock()
	defer m.unlock()

	for i := uint64(0); i < m.busySlots; i++ {
		idxValue := m.storage[i].filledIdx
		slot := m.storage[idxValue].mapSlot
		if !slot.isSet {
			return fmt.Errorf("!slot.isSet: %v: %v, %v", i, slot, idxValue)
		}
	}

	count := 0
	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if !slot.isSet {
			continue
		}

		count++
		fmt.Println("dump:", i, m.storage[i].mapSlot)
	}

	if count != m.Count() {
		return fmt.Errorf("count != m.Count(): %v %v", count, m.Count())
	}

	for i := uint64(0); i < m.size(); i++ {
		slot := m.storage[i].mapSlot
		if !slot.isSet {
			continue
		}

		idxValue := m.storage[slot.filledIdxIdx].filledIdx
		if i != idxValue {
			return fmt.Errorf("i != idxValue: i:%v idxValue:%v filledIdxIdx:%v imposterReverseFilledIdx:%v", i, idxValue, slot.filledIdxIdx, m.storage[m.storage[slot.filledIdxIdx].filledIdx].mapSlot.filledIdxIdx)
		}

		foundValue, err := m.Get(slot.key)
		if foundValue != slot.value || err != nil {
			hashValue := m.hashFunc(maximalSize, slot.key)
			expectedIdxValue := m.getIdx(hashValue)
			return fmt.Errorf("m.Get(slot.key) != slot.value: %v(%v) %v; i:%v key:%v expectedIdx:%v", foundValue, err, slot.value, i, slot.key, expectedIdxValue)
		}
	}

	return nil
}

func (m *openAddressGrowingMap) HasCollisionWithKey(key I.Key) bool {
	m.lock()
	defer m.unlock()

	hashValue := m.hashFunc(maximalSize, key)
	idxValue := m.getIdx(hashValue)

	return m.storage[idxValue].mapSlot.isSet
}
