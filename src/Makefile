
ifeq (${GOXGEN},1)
all: goxgen
else
all:
	@echo "Error: should run make at the repository root" 1>&2
	@exit 1
endif

goxgen:
	go build ./cmd/goxgen

.PONY: all
