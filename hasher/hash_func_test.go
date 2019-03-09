package hasher

import (
	"testing"

	benchmark "github.com/xaionaro-go/atomicmap/internal/benchmarkRoutines"
)

func TestHashCollisions_blockSize16_keyAmount16(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 16, 16)
}
func TestHashCollisions_blockSize64_keyAmount16(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 64, 16)
}
func TestHashCollisions_blockSize128_keyAmount16(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 128, 16)
}
func TestHashCollisions_blockSize1024_keyAmount16(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 1024, 16)
}

func TestHashCollisions_blockSize64_keyAmount64(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 64, 64)
}
func TestHashCollisions_blockSize128_keyAmount64(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 128, 64)
}
func TestHashCollisions_blockSize1024_keyAmount64(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 1024, 64)
}

func TestHashCollisions_blockSize1024_keyAmount380(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 1024, 380)
}
func TestHashCollisions_blockSize1024_keyAmount800(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 1024, 800)
}
func TestHashCollisions_blockSize1024_keyAmount1024(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 1024, 1024)
}

func TestHashCollisions_blockSize65536_keyAmount64(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 64)
}
func TestHashCollisions_blockSize65536_keyAmount380(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 380)
}
func TestHashCollisions_blockSize65536_keyAmount800(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 800)
}
func TestHashCollisions_blockSize65536_keyAmount1024(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 1024)
}
func TestHashCollisions_blockSize65536_keyAmount4096(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 4096)
}
func TestHashCollisions_blockSize65536_keyAmount32768(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 32768)
}
func TestHashCollisions_blockSize65536_keyAmount60000(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 60000)
}
func TestHashCollisions_blockSize65536_keyAmount65536(t *testing.T) {
	benchmark.DoTestHashCollisions(t, New(), 65536, 65536)
}

func BenchmarkHash_intKeyType_blockSize16(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 16, "int")
}
func BenchmarkHash_stringKeyType_blockSize16(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 16, "string")
}

func BenchmarkHash_intKeyType_blockSize64(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 64, "int")
}
func BenchmarkHash_stringKeyType_blockSize64(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 64, "string")
}

func BenchmarkHash_intKeyType_blockSize128(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 128, "int")
}
func BenchmarkHash_stringKeyType_blockSize128(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 128, "string")
}

func BenchmarkHash_intKeyType_blockSize1024(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 1024, "int")
}
func BenchmarkHash_stringKeyType_blockSize1024(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 1024, "string")
}

func BenchmarkHash_intKeyType_blockSize65536(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 65536, "int")
}
func BenchmarkHash_stringKeyType_blockSize65536(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 65536, "string")
}

func BenchmarkHash_intKeyType_blockSize1048576(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 1048576, "int")
}
func BenchmarkHash_stringKeyType_blockSize1048576(b *testing.B) {
	benchmark.DoBenchmarkHash(b, New(), 1048576, "string")
}
