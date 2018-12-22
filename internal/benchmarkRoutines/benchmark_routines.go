package benchmarkRoutines

import (
	"testing"

	I "github.com/xaionaro-go/atomicmap/interfaces"
)

func DoBenchmarkOfSet(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	keys := generateKeys(keyAmount, keyType)

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

	keys := generateKeys(keyAmount, keyType)
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

	keys := generateKeys(keyAmount, keyType)
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
func DoBenchmarkOfGetMiss(b *testing.B, factoryFunc mapFactoryFunc, customHasher I.Hasher, blockSize uint64, keyAmount uint64, keyType string) {
	b.StopTimer()
	b.ResetTimer()

	m := factoryFunc(blockSize, customHasher)

	otherKeyType := "string"
	if keyType == "string" {
		otherKeyType = "int"
	}
	otherKeys := generateKeys(keyAmount, otherKeyType)
	for i := uint64(0); i < keyAmount; i++ {
		m.Set(otherKeys[i], int(i))
	}

	keys := generateKeys(uint64(b.N), keyType)

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
	keys := generateKeys(keyAmount, keyType)

	currentIdx := uint64(0)
	if keyAmount >= 1024*1024 {
		b.ReportAllocs()
	}
	for i := 0; i < b.N; i++ {
		if currentIdx == 0 {
			b.StopTimer()
			for i := uint64(0); i < keyAmount; i++ {
				m.Set(keys[currentIdx], int(i))
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

	keys := generateKeys(uint64(b.N), keyType)

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

	keys := generateKeys(uint64(b.N), keyType)

	b.ReportAllocs()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		customHasher.Hash(blockSize, keys[i])
	}

	b.StopTimer()
}
