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
	@go test -run - -bench=. -benchmem -benchtime=125ms -count=10 | tee $(results)
	@sed '/^Benchmark/ { /^Benchmark.*\/\(os\|mmap\)/!d; }' $(results) | sed -e "/^Benchmark.*\/os/d" | sed -e "/^Benchmark/s|/mmap||g" > $(mmap)
	@sed '/^Benchmark/ { /^Benchmark.*\/\(os\|mmap\)/!d; }' $(results) | sed -e "/^Benchmark.*\/mmap/d" | sed -e "/^Benchmark/s|/os||g" > $(os)
