.PHONY: all test

# Set "-tags netgo" for really-really static compilation.
# See full explanations at: https://wjd.nu/notes/2016#golang-statically-linked
GOFLAGS = -tags netgo

all: test gospawn

test: $(shell find ! -path './.*' -name '*_test.go' -type f)
	-find ! -path '*/.*' -type d | xargs -n 1 go test -v

gospawn: Makefile $(shell find ! -path './.*' -name '*.go' -type f)
	-find ! -path './.*' -type d | xargs -n 1 golint
	@echo
	go build $(GOFLAGS) gospawn.go
	@if objdump -p gospawn | grep NEEDED; then \
	   echo '*** not a static binary ***' >&2; rm gospawn; false; \
	 fi
