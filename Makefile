.PHONY: all

# Set "-tags netgo" for really-really static compilation.
# See full explanations at: https://wjd.nu/notes/2016#golang-statically-linked
GOFLAGS = -tags netgo

all: gospawn

gospawn: Makefile $(shell find . -name '*.go' -type f)
	-find . -type d | xargs -n 1 golint
	@echo
	go build $(GOFLAGS) gospawn.go
