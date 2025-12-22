.PHONY: benchstat
benchstat:
	@rm -f benchstat benchstat.csv
	@go tool benchstat os.File mmapfile 2>/dev/null | tee benchstat
	@go tool benchstat -format=csv os.File mmapfile 2>/dev/null > benchstat.csv