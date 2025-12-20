.PHONY: benchstat
benchstat:
	@rm -f benchstat benchstat.csv
	@go tool benchstat os mmap 2>/dev/null | tee benchstat
	@go tool benchstat -format=csv os mmap 2>/dev/null > benchstat.csv