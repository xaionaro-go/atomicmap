package benchmarkRoutines

import (
	"testing"

	"git.dx.center/trafficstars/testJob0/internal/routines"
)

func DoBenchmarkOfSet(b *testing.B, factoryFunc mapFactoryFunc, blockSize uint32, keyAmount uint64, keyIsString bool) {
	b.StopTimer()

	m := factoryFunc(int(blockSize), routines.HashFunc)

	keys := generateKeys(keyAmount, keyIsString)

	currentCount := uint64(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(keys[currentCount], i)
		currentCount++
		if currentCount >= keyAmount {
			b.StopTimer()
			m = factoryFunc(int(blockSize), routines.HashFunc)
			currentCount = 0
			b.StartTimer()
		}
	}
	b.StopTimer()
}

func DoBenchmarkOfReSet(b *testing.B, factoryFunc mapFactoryFunc, blockSize uint32, keyAmount uint64, keyIsString bool) {
	b.StopTimer()

	m := factoryFunc(int(blockSize), routines.HashFunc)

	keys := generateKeys(keyAmount, keyIsString)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], i+1)
	}

	currentIdx := uint64(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(keys[currentIdx], i)
		currentIdx++
		if currentIdx >= keyAmount {
			currentIdx = 0
		}
	}
	b.StopTimer()
}
func DoBenchmarkOfGet(b *testing.B, factoryFunc mapFactoryFunc, blockSize uint32, keyAmount uint64, keyIsString bool) {
	b.StopTimer()

	m := factoryFunc(int(blockSize), routines.HashFunc)

	keys := generateKeys(keyAmount, keyIsString)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], i)
	}

	currentIdx := uint64(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Get(keys[currentIdx])
		currentIdx++
		if currentIdx >= keyAmount {
			currentIdx = 0
		}
	}
	b.StopTimer()
}
func DoBenchmarkOfGetMiss(b *testing.B, factoryFunc mapFactoryFunc, blockSize uint32, keyAmount uint64, keyIsString bool) {
	b.StopTimer()

	m := factoryFunc(int(blockSize), routines.HashFunc)

	keys := generateKeys(uint64(b.N), keyIsString)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Get(keys[i])
	}
	b.StopTimer()
}
func DoBenchmarkOfUnset(b *testing.B, factoryFunc mapFactoryFunc, blockSize uint32, keyAmount uint64, keyIsString bool) {
	b.StopTimer()

	m := factoryFunc(int(blockSize), routines.HashFunc)
	keys := generateKeys(keyAmount, keyIsString)

	currentIdx := uint64(0)
	for i := 0; i < b.N; i++ {
		if currentIdx == 0 {
			b.StopTimer()
			for i := uint64(0); i < keyAmount; i++ {
				m.Set(keys[currentIdx], i)
			}
			b.StartTimer()
		}

		m.Unset(keys[currentIdx])

		currentIdx++
		if currentIdx >= keyAmount {
			currentIdx = 0
		}
	}
	b.StopTimer()
}
func DoBenchmarkOfUnsetMiss(b *testing.B, factoryFunc mapFactoryFunc, blockSize uint32, keyAmount uint64, keyIsString bool) {
	b.StopTimer()

	m := factoryFunc(int(blockSize), routines.HashFunc)

	keys := generateKeys(uint64(b.N), keyIsString)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Unset(keys[i])
	}
	b.StopTimer()
}
