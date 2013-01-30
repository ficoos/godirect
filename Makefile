GOPATH := $(CURDIR)

GO_INSTALL := go install
GO_BUILD := go build
GO_TEST := go test
GO_CLEAN := go clean
GO_GET := go get

GO_DEPS := launchpad.net/gocheck

.PHONY: godirect get_deps update_deps

all: get_deps bin/godirect

get_deps:
	GOPATH="$(GOPATH)" ${GO_GET} ${GO_DEPS}

update_deps:
	GOPATH="$(GOPATH)" ${GO_GET} -u ${GO_DEPS}


bin/godirect: godirect
	GOPATH="$(GOPATH)" ${GO_INSTALL} $<
	GOPATH="$(GOPATH)" ${GO_TEST} $<

clean:
	GOPATH="$(GOPATH)" ${GO_CLEAN} godirect
	rm -f bin/*

