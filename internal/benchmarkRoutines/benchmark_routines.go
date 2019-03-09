package benchmarkRoutines

import (
	"sync/atomic"
	"testing"

	I "github.com/xaionaro-go/atomicmap/interfaces"
)

func DoBenchmarkOfSet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, keyType)

	currentCount := uint64(0)
	if keyAmount >= 1024*1024 {
		b.ReportAllocs()
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(keys[currentCount], i)
		currentCount++
		if currentCount >= keyAmount {
			b.StopTimer()
			m = factoryFunc(blockSize, customHasher)
			currentCount = 0
			b.StartTimer()
		}
	}
	b.StopTimer()
}

func DoBenchmarkOfReSet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, keyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], int(i+1))
	}

	currentIdx := uint64(0)
	b.ReportAllocs()
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
func DoBenchmarkOfGet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, keyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], int(i))
	}

	currentIdx := uint64(0)
	b.ReportAllocs()
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
func DoBenchmarkOfGetByBytes(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, `bytes`)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], int(i))
	}

	currentIdx := uint64(0)
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.GetByBytes(keys[currentIdx].([]byte))
		currentIdx++
		if currentIdx >= keyAmount {
			currentIdx = 0
		}
	}
	b.StopTimer()
}
func DoBenchmarkOfGetMiss(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	otherKeyType := "string"
	if keyType == "string" {
		otherKeyType = "int"
	}
	otherKeys := generateKeys(customHasher, keyAmount, otherKeyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(otherKeys[i], int(i))
	}

	keys := generateKeys(customHasher, uint64(b.N), keyType)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Get(keys[i])
	}
	b.StopTimer()
}
func DoBenchmarkOfUnset(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)
	keys := generateKeys(customHasher, keyAmount, keyType)

	currentIdx := uint64(0)
	if keyAmount >= 1024*1024 {
		b.ReportAllocs()
	}
	for i := 0; i < b.N; i++ {
		if currentIdx == 0 {
			b.StopTimer()
			for i := uint64(0); i < keyAmount; i++ {
				m.Set(keys[i], int(i))
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
func DoBenchmarkOfUnsetMiss(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, uint64(b.N), keyType)

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Unset(keys[i])
	}
	b.StopTimer()
}

func DoBenchmarkHash(b *testing.B, customHasher I.Hasher, blockSize uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	keys := generateKeys(customHasher, uint64(b.N), keyType)

	b.ReportAllocs()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		customHasher.CompressHash(blockSize, customHasher.Hash(keys[i]))
	}

	b.StopTimer()
}

func DoParallelBenchmarkOfSet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, keyType)

	currentCount := uint64(0)
	if keyAmount >= 1024*1024 {
		b.ReportAllocs()
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			localCurrentCount := atomic.AddUint64(&currentCount, 1)
			if localCurrentCount >= keyAmount {
				b.StopTimer()
				m = factoryFunc(blockSize, customHasher)
				localCurrentCount = 0
				atomic.StoreUint64(&currentCount, 0)
				b.StartTimer()
			}
			m.Set(keys[localCurrentCount], localCurrentCount)
		}
	})
	b.StopTimer()
}

func DoParallelBenchmarkOfReSet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, keyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], int(i+1))
	}

	currentIdx := uint64(0)
	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			localCurrentIdx := atomic.AddUint64(&currentIdx, 1)
			if localCurrentIdx >= keyAmount {
				localCurrentIdx = 0
				atomic.StoreUint64(&currentIdx, 0)
			}
			m.Set(keys[localCurrentIdx], localCurrentIdx)
		}
	})
	b.StopTimer()
}
func DoParallelBenchmarkOfGet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, keyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], int(i))
	}

	currentIdx := uint64(0)
	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			localCurrentIdx := atomic.AddUint64(&currentIdx, 1)
			if localCurrentIdx >= keyAmount {
				localCurrentIdx = 0
				atomic.StoreUint64(&currentIdx, 0)
			}
			m.Get(keys[localCurrentIdx])
		}
	})
	b.StopTimer()
}
func DoParallelBenchmarkOfGetByBytes(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, keyAmount, `bytes`)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(keys[i], int(i))
	}

	currentIdx := uint64(0)
	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			localCurrentIdx := atomic.AddUint64(&currentIdx, 1)
			if localCurrentIdx >= keyAmount {
				localCurrentIdx = 0
				atomic.StoreUint64(&currentIdx, 0)
			}
			m.GetByBytes(keys[localCurrentIdx].([]byte))
		}
	})
	b.StopTimer()
}
func DoParallelBenchmarkOfGetMiss(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	otherKeyType := "string"
	if keyType == "string" {
		otherKeyType = "int"
	}
	otherKeys := generateKeys(customHasher, keyAmount, otherKeyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(otherKeys[i], int(i))
	}

	currentIdx := uint64(0)
	keys := generateKeys(customHasher, uint64(b.N), keyType)

	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			localCurrentIdx := atomic.AddUint64(&currentIdx, 1)
			if localCurrentIdx >= keyAmount {
				localCurrentIdx = 0
				atomic.StoreUint64(&currentIdx, 0)
			}
			m.Get(keys[localCurrentIdx])
		}
	})
	b.StopTimer()
}

func DoParallelBenchmarkOfUnsetMiss(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(customHasher, uint64(b.N), keyType)

	b.ReportAllocs()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Unset(keys[0])
		}
	})
	b.StopTimer()
}
