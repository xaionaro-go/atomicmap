// This file had been automatically generated by utility "git.dx.center/trafficstars/testJob0/internal/benchmarkCodeGen"

package stupidMap

import (
	"testing"

	"git.dx.center/trafficstars/testJob0/internal/routines"
	benchmark "git.dx.center/trafficstars/testJob0/internal/benchmarkRoutines"
)

func Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 16, "int")
}

func Benchmark_stupidMap_Set_stringKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 16, "string")
}

func Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 512, "int")
}

func Benchmark_stupidMap_Set_stringKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 512, "string")
}

func Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 65536, "int")
}

func Benchmark_stupidMap_Set_stringKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 65536, "string")
}

func Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 1048576, "int")
}

func Benchmark_stupidMap_Set_stringKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfSet(b, NewHashMap, routines.HashFunc, 0, 1048576, "string")
}

func Benchmark_stupidMap_ReSet_intKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 16, "int")
}

func Benchmark_stupidMap_ReSet_stringKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 16, "string")
}

func Benchmark_stupidMap_ReSet_intKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 512, "int")
}

func Benchmark_stupidMap_ReSet_stringKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 512, "string")
}

func Benchmark_stupidMap_ReSet_intKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 65536, "int")
}

func Benchmark_stupidMap_ReSet_stringKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 65536, "string")
}

func Benchmark_stupidMap_ReSet_intKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 1048576, "int")
}

func Benchmark_stupidMap_ReSet_stringKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfReSet(b, NewHashMap, routines.HashFunc, 0, 1048576, "string")
}

func Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 16, "int")
}

func Benchmark_stupidMap_Get_stringKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 16, "string")
}

func Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 512, "int")
}

func Benchmark_stupidMap_Get_stringKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 512, "string")
}

func Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 65536, "int")
}

func Benchmark_stupidMap_Get_stringKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 65536, "string")
}

func Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 1048576, "int")
}

func Benchmark_stupidMap_Get_stringKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfGet(b, NewHashMap, routines.HashFunc, 0, 1048576, "string")
}

func Benchmark_stupidMap_GetMiss_intKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 16, "int")
}

func Benchmark_stupidMap_GetMiss_stringKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 16, "string")
}

func Benchmark_stupidMap_GetMiss_intKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 512, "int")
}

func Benchmark_stupidMap_GetMiss_stringKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 512, "string")
}

func Benchmark_stupidMap_GetMiss_intKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 65536, "int")
}

func Benchmark_stupidMap_GetMiss_stringKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 65536, "string")
}

func Benchmark_stupidMap_GetMiss_intKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 1048576, "int")
}

func Benchmark_stupidMap_GetMiss_stringKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfGetMiss(b, NewHashMap, routines.HashFunc, 0, 1048576, "string")
}

func Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 16, "int")
}

func Benchmark_stupidMap_Unset_stringKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 16, "string")
}

func Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 512, "int")
}

func Benchmark_stupidMap_Unset_stringKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 512, "string")
}

func Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 65536, "int")
}

func Benchmark_stupidMap_Unset_stringKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 65536, "string")
}

func Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 1048576, "int")
}

func Benchmark_stupidMap_Unset_stringKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfUnset(b, NewHashMap, routines.HashFunc, 0, 1048576, "string")
}

func Benchmark_stupidMap_UnsetMiss_intKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 16, "int")
}

func Benchmark_stupidMap_UnsetMiss_stringKeyType_blockSize0_keyAmount16(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 16, "string")
}

func Benchmark_stupidMap_UnsetMiss_intKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 512, "int")
}

func Benchmark_stupidMap_UnsetMiss_stringKeyType_blockSize0_keyAmount512(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 512, "string")
}

func Benchmark_stupidMap_UnsetMiss_intKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 65536, "int")
}

func Benchmark_stupidMap_UnsetMiss_stringKeyType_blockSize0_keyAmount65536(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 65536, "string")
}

func Benchmark_stupidMap_UnsetMiss_intKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 1048576, "int")
}

func Benchmark_stupidMap_UnsetMiss_stringKeyType_blockSize0_keyAmount1048576(b *testing.B) {
	benchmark.DoBenchmarkOfUnsetMiss(b, NewHashMap, routines.HashFunc, 0, 1048576, "string")
}
