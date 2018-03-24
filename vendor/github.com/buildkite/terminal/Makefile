SRC=*.go cmd/terminal-to-html/*.go
BINARY=terminal-to-html
BUILDCMD=godep go build -o $@ cmd/terminal-to-html/*
VERSION=$(shell cat version.go  | grep baseVersion | head -n1 | cut -d \" -f 2)

all: test $(BINARY)

bench:
	godep go test -bench . -benchmem

test:
	godep go test

clean:
	rm -f $(BINARY)
	rm -rf dist bin

cmd/terminal-to-html/_bindata.go: assets/terminal.css
	go-bindata -o cmd/terminal-to-html/bindata.go -nomemcopy assets

$(BINARY): $(SRC)
	$(BUILDCMD)

version:
	@echo $(VERSION)

# Cross-compiling

GZ_ARCH     := linux-amd64 linux-386 linux-arm darwin-386 darwin-amd64
ZIP_ARCH    := windows-386 windows-amd64
GZ_TARGETS  := $(foreach tar,$(GZ_ARCH), dist/$(BINARY)-$(VERSION)-$(tar).gz)
ZIP_TARGETS := $(foreach tar,$(ZIP_ARCH), dist/$(BINARY)-$(VERSION)-$(tar).zip)

dist: $(GZ_TARGETS) $(ZIP_TARGETS)

dist/%.gz: bin/%
	@[ -d dist ] || mkdir dist
	gzip -c $< > $@

dist/%.zip: bin/%
	@[ -d dist ] || mkdir dist
	@rm -f $@ || true
	zip $@ $<

bin/$(BINARY)-$(VERSION)-%: $(SRC)
	@[ -d bin ] || mkdir bin
	GOOS=$(firstword $(subst -, , $*)) GOARCH=$(lastword $(subst -, , $*)) $(BUILDCMD)

.PHONY: clean bench test dist version
