
GOROOT="$(shell pwd)/go"
GO_PATH="${GOROOT}/bin"
GOXGEN="1"

all: go goxgen

go:
	cd go/src && ./make.bash

goxgen:
	GOXGEN="${GOXGEN}" GOROOT="${GOROOT}" make -C ./src

env:
	@echo "export GOROOT='${GOROOT}'"

.PHONY: all go goxgen

