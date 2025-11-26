.DEFAULT_GOAL := help

go-xray:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o go-xray

clean:
	rm -f go-xray

help:
	@echo '---    help    ---'
	@echo 'make go-xray - compile application for current running architecture'
	@echo 'make clean   - clean compilation artifacts'
	@echo '--- good luck! ---'
