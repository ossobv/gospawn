.PHONY: all test

# Set "-tags netgo" for really-really static compilation.
# See full explanations at: https://wjd.nu/notes/2016#golang-statically-linked
GOFLAGS = -tags netgo

all: test gospawn

test: $(shell find ! -path './.*' -type f -name '*_test.go')
	-find ! -path '*/.*' -type d | xargs -n 1 go test -v
	-find ! -path '*/.*' -type f -name '*.go' | xargs -n 1 gofmt -s -d

gospawn: Makefile $(shell find ! -path './.*' -name '*.go' -type f)
	-find ! -path './.*' -type d | xargs -n 1 golint
	@echo
	go build $(GOFLAGS) \
	  -ldflags "-X main.version=$(shell git describe --always --dirty)" \
	  gospawn.go
	@if objdump -p gospawn | grep NEEDED; then \
	   echo '*** not a static binary ***' >&2; rm gospawn; false; \
	fi

gospawn.upx: gospawn
	if cp -a gospawn gospawn.upx && upx -qq gospawn.upx; then \
	  touch gospawn.upx; else rm gospawn.upx; fi
