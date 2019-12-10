BINARY_NAME := wg-connect
PLATFORMS := linux/amd64 darwin/amd64 windows/amd64
GO := go
TEMP = $(subst /, ,$@)
OS = $(word 1, $(TEMP))
ARCH = $(word 2, $(TEMP))

.PHONY: all
all: test build

.PHONY: build
build: $(PLATFORMS)

$(PLATFORMS):
				GOOS=$(OS) GOARCH=$(arch) $(GO) build -o 'bin/$(BINARY_NAME)-$(OS)-$(ARCH)'

.PHONY: test
test:
				$(GO) test -v ./...

.PHONY: clean
clean:
				$(GO) clean
				rm -f bin/$(BINARY_NAME)-*