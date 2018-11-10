```
Hash function:

Total collisions on random keys: collisions 66, keyAmount 380 and blockSize 1024:
        66/380/1024 (17.4%)
Total collisions on keys of pessimistic scenario (keys are multiple of blockSize): collisions 63, keyAmount 380 and blockSize 1024:
        63/380/1024 (16.6%)
Total collisions on keys of pessimistic scenario (keys are consecutive): collisions 0, keyAmount 380 and blockSize 1024:
        0/380/1024 (0.0%)

BenchmarkHash_intKeyType_blockSize16-8                  10000000                21.5 ns/op             0 B/op          0 allocs/op
BenchmarkHash_stringKeyType_blockSize16-8                2000000               105 ns/op               0 B/op          0 allocs/op
BenchmarkHash_intKeyType_blockSize1048576-8             10000000                17.0 ns/op             0 B/op          0 allocs/op
BenchmarkHash_stringKeyType_blockSize1048576-8           2000000                98.7 ns/op             0 B/op          0 allocs/op


my hash-map:

Set:
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize16_keyAmount1048576-8                     500000              1076 ns/op             343 B/op          1 allocs/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize16_keyAmount16-8                          500000               297 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize64_keyAmount16-8                         1000000               149 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize128_keyAmount16-8                        1000000               151 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize1024_keyAmount16-8                       1000000               175 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize1024_keyAmount512-8                      1000000               113 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize65536_keyAmount512-8                     1000000               136 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize65536_keyAmount65536-8                   1000000               459 ns/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize4194304_keyAmount1048576-8               1000000               150 ns/op               8 B/op          0 allocs/op
Benchmark_openAddressGrowingMap_Set_intKeyType_blockSize16777216_keyAmount1048576-8              1000000               150 ns/op               8 B/op          0 allocs/op

Get:
Benchmark_openAddressGrowingMap_Get_intKeyType_blockSize1024_keyAmount16-8                       5000000                37.8 ns/op             0 B/op          0 allocs/op
Benchmark_openAddressGrowingMap_Get_intKeyType_blockSize1024_keyAmount512-8                      3000000                47.4 ns/op             0 B/op          0 allocs/op
Benchmark_openAddressGrowingMap_Get_intKeyType_blockSize1024_keyAmount65536-8                    2000000                92.3 ns/op             0 B/op          0 allocs/op
Benchmark_openAddressGrowingMap_Get_intKeyType_blockSize1024_keyAmount1048576-8                  1000000               130 ns/op               0 B/op          0 allocs/op

Unset:
Benchmark_openAddressGrowingMap_Unset_intKeyType_blockSize1024_keyAmount16-8                     5000000                35.5 ns/op
Benchmark_openAddressGrowingMap_Unset_intKeyType_blockSize1024_keyAmount512-8                   20000000                 7.71 ns/op
Benchmark_openAddressGrowingMap_Unset_intKeyType_blockSize1024_keyAmount65536-8                 20000000                 6.72 ns/op
Benchmark_openAddressGrowingMap_Unset_intKeyType_blockSize1024_keyAmount1048576-8               30000000                 6.04 ns/op            0 B/op          0 allocs/op


native map:

Set:
Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount16-8                       300000               470 ns/op
Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount512-8                      500000               363 ns/op
Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount65536-8                    300000               490 ns/op
Benchmark_stupidMap_Set_intKeyType_blockSize0_keyAmount1048576-8                  500000               653 ns/op             170 B/op          1 allocs/op

Get:
Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount16-8                      3000000                54.2 ns/op             0 B/op          0 allocs/op
Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount512-8                     2000000                61.9 ns/op             0 B/op          0 allocs/op
Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount65536-8                   2000000                78.0 ns/op             0 B/op          0 allocs/op
Benchmark_stupidMap_Get_intKeyType_blockSize0_keyAmount1048576-8                 1000000               137 ns/op               0 B/op          0 allocs/op

Unset:
Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount16-8                    3000000                50.0 ns/op
Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount512-8                  20000000                12.3 ns/op
Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount65536-8                20000000                11.1 ns/op
Benchmark_stupidMap_Unset_intKeyType_blockSize0_keyAmount1048576-8              20000000                10.2 ns/op             0 B/op          0 allocs/op
```
