Please test the result on a multicore machine (>6 cores)

See also solutions/comments

 -- Dmitry Yu Okunev, 2018-07-16

I considered that the "block size" is the initial number of elements of the
internal storage of the map. The reason:

 * "block size" cannot be the size of index value (even in bits) because
   nowhere exists amounts of memory 2^1024 (you require to make a benchmark
   for blocksize 1024).

 * "block size" is unlikely to be the static size of the internal storage of
   the map because it's unlikely to have any interest to map of mush less of
   16 elements (you require to make a benchmark for blocksize 16 and there's
   no sence to use highly loaded map, so the real number of elements should be
   much less than even 16).

 * "block size" is unlikely to be the number of elements to grow the internal
   storage because you require to implements the minimal amount of memory
   allocations and the common behaviour in the case is doubling the allocated
   block (instead of a linear growing by a some fixed "block size").

 -- Dmitry Yu Okunev, 2018-07-14
