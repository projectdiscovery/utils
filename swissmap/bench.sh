ITEMS=0
if [ "$1" != "" ]; then
    ITEMS=$1
fi
rm -rf *bench.out
go test -run - -count=10 -timeout=1h -benchmem -bench . -items="$ITEMS" | \
    tee bench.out
grep -v "swissmap" bench.out | grep -v "BenchmarkHelper" | sed "s|/impl=mapsutil||g" > mapsutil-bench.out
grep -v "mapsutil" bench.out | grep -v "BenchmarkHelper" | sed "s|/impl=swissmap||g" > swissmap-bench.out
benchstat swissmap-bench.out mapsutil-bench.out