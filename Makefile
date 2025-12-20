benchdata := benchdata
results := $(benchdata)/results
mmap := $(benchdata)/mmap
os := $(benchdata)/os

.PHONY: all
all: test bench

.PHONY: test
test:
	@go test -v -race .

.PHONY: bench
bench:
	@mkdir $(benchdata) 2>/dev/null || true
	@rm -rf $(results)
	@go test -run - -bench=. -benchmem -count=10 | tee $(results)
	@sed -e "/^Benchmark.*\/os/d" $(results) | sed -e "/^Benchmark/s|/mmap||g" > $(mmap)
	@sed -e "/^Benchmark.*\/mmap/d" $(results) | sed -e "/^Benchmark/s|/os||g" > $(os)
