Task: Implement HashMap data structure

Demands
-------

* Creation of new HashMap object like `map = new HashMap(124 / block size /, [hashFunction] / optional /)`
* Key is `string` or *dynamic_type*.
* Value is `dynamic_type`
* Add test/benchmark for `HashMap(Set, Get, Unset)` with size of block 16, 64, 128, 1024. Number of memory allocations, time spent for 1000000 operations.
* Add test/benchmark for Hash function with size of block 16, 64, 128, 1024. Number of memory allocations and collisions.
* Compare your implementation with native `map` type

Whats important
---------------

* Memory allocations
* Performance of operations

Interfaces
----------

See `./interfaces/`

The factory function should be of type `func NewHashMap(blockSize int, fn func(blockSize int, key Key) int) interfaces.HashMaper`
