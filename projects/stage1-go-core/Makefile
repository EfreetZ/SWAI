.PHONY: test race bench perf stress chaos fmt vet clean

test:
	go test ./... -v -count=1

race:
	go test -race ./... -v

bench:
	go test -bench=. -benchmem ./...

perf:
	go test ./test/performance -v -count=1

stress:
	go test ./test/stress -v -count=1

chaos:
	go test ./test/chaos -v -count=1

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	go clean -testcache
