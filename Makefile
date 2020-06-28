
GOROOT="${PWD}/go"
GO_PATH="${GOROOT}/bin"
GOXGEN="1"

all: go goxgen

go:
	cd go/src && ./make.bash

goxgen:
	GOXGEN="${GOXGEN}" GOROOT="${GOROOT}" make -C ./src

env:
	@echo "export GOROOT='${GOROOT}'"
	@echo "${PATH}" | grep -q $(GO_PATH) && true || echo 'export PATH=$${PATH}:'$(GO_PATH)

.PHONY: all go goxgen

