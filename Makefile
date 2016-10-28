GO ?= go
export GOPATH := $(CURDIR)/_vendor

logcarrier-storage:
	$(GO) build storage/logcarrier-storage.go

submodules:
	git submodule init
	git submodule update --recursive

clean:
	rm -fv {.,storage}/logcarrier-storage
