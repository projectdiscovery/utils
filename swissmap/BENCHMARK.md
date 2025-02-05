# Benchmark

* **swissmap** *(uses [cockroachdb/swiss](https://github.com/cockroachdb/swiss) under the hood)*

```
BenchmarkGet/items=1000/seq-16                           8000854 		139.2 ns/op		12  B/op	1 allocs/op
BenchmarkGet/items=1000/parallel-16                      49186312		24.73 ns/op		12  B/op	1 allocs/op
BenchmarkGet/items=5000/seq-16                           8156024 		135.9 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=5000/parallel-16                      45306098		26.74 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=10000/seq-16                          8093172 		140.4 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=10000/parallel-16                     44395874		31.18 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=100000/seq-16                         6251833 		189.4 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=100000/parallel-16                    39724492		31.15 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=250000/seq-16                         5020261 		222.7 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=250000/parallel-16                    49155284		30.00 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=500000/seq-16                         4082534 		288.1 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=500000/parallel-16                    42062390		29.94 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000000/seq-16                        3348501 		354.1 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000000/parallel-16                   43152501		31.64 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000/WithThreadSafety/seq-16          8574296 		139.3 ns/op		12  B/op	1 allocs/op
BenchmarkGet/items=1000/WithThreadSafety/parallel-16     10291254		138.4 ns/op		12  B/op	1 allocs/op
BenchmarkGet/items=5000/WithThreadSafety/seq-16          7859248 		146.9 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=5000/WithThreadSafety/parallel-16     10370797		138.2 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=10000/WithThreadSafety/seq-16         7797219 		151.9 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=10000/WithThreadSafety/parallel-16    9954980 		143.8 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=100000/WithThreadSafety/seq-16        5443500 		208.6 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=100000/WithThreadSafety/parallel-16   9368031 		142.4 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=250000/WithThreadSafety/seq-16        4688386 		277.2 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=250000/WithThreadSafety/parallel-16   9505594 		138.6 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=500000/WithThreadSafety/seq-16        3644149 		320.5 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=500000/WithThreadSafety/parallel-16   8204418 		140.5 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000000/WithThreadSafety/seq-16       3098900 		382.3 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000000/WithThreadSafety/parallel-16  8717754 		138.7 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=1000/WithThreadSafety/seq-16          2570617 		466.2 ns/op		37  B/op	2 allocs/op
BenchmarkSet/items=1000/WithThreadSafety/parallel-16     2163002 		548.9 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=5000/WithThreadSafety/seq-16          2594163 		465.3 ns/op		37  B/op	2 allocs/op
BenchmarkSet/items=5000/WithThreadSafety/parallel-16     2159650 		529.8 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=10000/WithThreadSafety/seq-16         2610625 		457.1 ns/op		36  B/op	2 allocs/op
BenchmarkSet/items=10000/WithThreadSafety/parallel-16    2265956 		526.8 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=100000/WithThreadSafety/seq-16        2556045 		482.4 ns/op		37  B/op	2 allocs/op
BenchmarkSet/items=100000/WithThreadSafety/parallel-16   2196990 		541.3 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=250000/WithThreadSafety/seq-16        2680706 		471.0 ns/op		36  B/op	2 allocs/op
BenchmarkSet/items=250000/WithThreadSafety/parallel-16   2444006 		509.2 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=500000/WithThreadSafety/seq-16        2899224 		452.2 ns/op		34  B/op	2 allocs/op
BenchmarkSet/items=500000/WithThreadSafety/parallel-16   2279362 		525.6 ns/op		15  B/op	1 allocs/op
BenchmarkSet/items=1000000/WithThreadSafety/seq-16       3387404 		463.5 ns/op		32  B/op	2 allocs/op
BenchmarkSet/items=1000000/WithThreadSafety/parallel-16  2167636 		546.3 ns/op		15  B/op	1 allocs/op
```

* mapsutil (`4a4cbd9`)

```
BenchmarkGet/items=1000/seq-16                           8214813 		144.5 ns/op		12  B/op	1 allocs/op
BenchmarkGet/items=1000/parallel-16                      52135828		24.30 ns/op		12  B/op	1 allocs/op
BenchmarkGet/items=5000/seq-16                           7584444 		157.0 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=5000/parallel-16                      47813564		26.42 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=10000/seq-16                          7216506 		160.1 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=10000/parallel-16                     46165209		28.49 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=100000/seq-16                         6186141 		195.1 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=100000/parallel-16                    39932752		29.89 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=250000/seq-16                         5442273 		255.5 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=250000/parallel-16                    51295606		28.32 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=500000/seq-16                         4219646 		263.2 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=500000/parallel-16                    50865244		28.94 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000000/seq-16                        4100514 		298.2 ns/op		15  B/op	1 allocs/op
BenchmarkGet/items=1000000/parallel-16                   49751544		30.14 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=1000/seq-16                    7747381 		155.3 ns/op		12  B/op	1 allocs/op
BenchmarkSyncMapGet/items=1000/parallel-16               13463876		98.39 ns/op		12  B/op	1 allocs/op
BenchmarkSyncMapGet/items=5000/seq-16                    6824139 		174.3 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=5000/parallel-16               10122182		120.5 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=10000/seq-16                   6453328 		173.2 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=10000/parallel-16              13732112		102.1 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=100000/seq-16                  5495906 		212.5 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=100000/parallel-16             11262956		119.9 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=250000/seq-16                  4041955 		289.1 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=250000/parallel-16             10759794		121.0 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=500000/seq-16                  3405672 		325.4 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=500000/parallel-16             13332009		97.60 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=1000000/seq-16                 3500380 		346.9 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapGet/items=1000000/parallel-16            11079632		120.0 ns/op		15  B/op	1 allocs/op
BenchmarkSyncMapSet/items=1000/seq-16                    1902699 		646.8 ns/op		146 B/op	2 allocs/op
BenchmarkSyncMapSet/items=1000/parallel-16               2409667 		497.8 ns/op		22  B/op	2 allocs/op
BenchmarkSyncMapSet/items=5000/seq-16                    1929216 		657.0 ns/op		144 B/op	2 allocs/op
BenchmarkSyncMapSet/items=5000/parallel-16               2402744 		504.2 ns/op		22  B/op	2 allocs/op
BenchmarkSyncMapSet/items=10000/seq-16                   2019657 		638.3 ns/op		138 B/op	2 allocs/op
BenchmarkSyncMapSet/items=10000/parallel-16              2319279 		527.0 ns/op		22  B/op	2 allocs/op
BenchmarkSyncMapSet/items=100000/seq-16                  1825004 		651.8 ns/op		151 B/op	2 allocs/op
BenchmarkSyncMapSet/items=100000/parallel-16             2501198 		549.2 ns/op		22  B/op	2 allocs/op
BenchmarkSyncMapSet/items=250000/seq-16                  1939453 		640.4 ns/op		143 B/op	2 allocs/op
BenchmarkSyncMapSet/items=250000/parallel-16             2542819 		546.9 ns/op		22  B/op	2 allocs/op
BenchmarkSyncMapSet/items=500000/seq-16                  1949194 		742.8 ns/op		143 B/op	2 allocs/op
BenchmarkSyncMapSet/items=500000/parallel-16             2523684 		572.5 ns/op		22  B/op	2 allocs/op
BenchmarkSyncMapSet/items=1000000/seq-16                 1987189 		720.9 ns/op		140 B/op	2 allocs/op
BenchmarkSyncMapSet/items=1000000/parallel-16            2249832 		507.2 ns/op		22  B/op	2 allocs/op
```

---

## key observations

### perf

* `Get` op
  * `swissmap` is faster than `mapsutil` for small-to-medium datasets in sequence
    * e.g., **139–151 ns/op** vs. **155–346 ns/op** at 1K–1M items
  * `mapsutil` outperforms `swissmap` in parallel
    * e.g., **98–120 ns/op** vs. **138–143 ns/op** for 1M items
* `Set` op
  * `mapsutil` is slightly faster than `swissmap` in parallel but uses more memory
    * `mapsutil`
      * **497–572 ns/op**
      * **22 B/op**
    * `swissmap`
      * **509–548 ns/op**
      * **15 B/op**
  * `swissmap` is significantly faster in sequence and *far more memory-efficient*
    * `mapsutil`
      * **497–572 ns/op**
      * **138–151 B/op**
    * `swissmap`
      * **450–550 ns/op**
      * **32–37 B/op**
* memory efficient
  * `swissmap` uses **~15 B/op** for `Get` and **15–37 B/op** for `Set` *(lower in parallel)*
  * `mapsutil` uses **12–15 B/op** for `Get` but **22–151 B/op** for `Set` (overhead in sequential `Set`)
* scalability
  * `swissmap` maintains stable parallel `Get` perf regardless of item count (**~138–143 ns/op**)

### conclusion

1. use `mapsutil` if parallel `Get` perf is critical and memory overhead is *"acceptable"*
1. prefer `swissmap` for seq workloads, memory-sensitive apps *(especially `Set`)*, and consistent parallel `Get` perf across large datasets
1. `swissmap` trades slightly slower parallel `Get` for better memory efficiency in `Set`, while `mapsutil` prioritizes speed in concurrent r/w

`mapsutil` for read-heavy, while `swissmap` for write-heavy scenarios or when memory efficiency is critical