.PHONY: test
test:
	go test -race -covermode=atomic -v -coverprofile=coverage.txt ./...

.PHONY: bench
bench:
	go test -bench=. -run=^Benchmark ./...