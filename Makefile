GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
Minversion := $(shell date)
BUILD_NODE_PAR = -ldflags "-X github.com/elastos/Elastos.ELA.SideChain/config.Version=$(VERSION)" #-race

all:
	$(GC)  $(BUILD_NODE_PAR) -o side main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6
