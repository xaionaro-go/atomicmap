all:
	cd $(go env GOPATH)/src/git.dx.center/trafficstars/testJob0
	go test -bench=. -benchtime 100ms -timeout 2h ./... | grep -E '^Benchmark_.*/op' | grep -v 'SIG\|trafficstars' | sed -r -e "s/[-_\t]/ /g" -e 's/[ ]{1,}/,/g' -e 's/op,.*/op/g' -e 's/KeyType//g' -e 's/blockSize//g' -e 's/keyAmount//g' > /tmp/bench_raw_results.csv
	go run post_process/prepare_csv/main.go /tmp/bench_raw_results > /tmp/bench_raw_results_postproc.csv
