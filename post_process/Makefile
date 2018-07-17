all:
	go test -bench=. -benchtime 100ms -timeout 2h ../... | tee /tmp/bench_rawraw_results.csv | grep --line-buffered -E '^Benchmark_.*/op' | grep --line-buffered -v 'SIG\|trafficstars' | sed -r -e "s/[-_\t]/ /g" -e 's/[ ]{1,}/,/g' -e 's/op,.*/op/g' -e 's/KeyType//g' -e 's/blockSize//g' -e 's/keyAmount//g' -e 's/ThreadSafety//g' | tee /tmp/bench_raw_results.csv
	go run prepare_csv/main.go /tmp/bench_raw_results.csv int Set > /tmp/bench_raw_results_int_set.csv
	go run prepare_csv/main.go /tmp/bench_raw_results.csv int Get > /tmp/bench_raw_results_int_get.csv
	go run prepare_csv/main.go /tmp/bench_raw_results.csv string Set > /tmp/bench_raw_results_string_set.csv
	go run prepare_csv/main.go /tmp/bench_raw_results.csv int Unset > /tmp/bench_raw_results_int_unset.csv
