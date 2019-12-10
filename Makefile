BINARY_NAME := wg-connect

.PHONY: all
all: linux darwin windows

linux:
	GOOS=linux GOARCH=amd64 go build -o 'bin/$(BINARY_NAME)-linux-amd64'

darwin:
	GOOS=darwin GOARCH=amd64 go build -o 'bin/$(BINARY_NAME)-darwin-amd64'

windows:
	GOOS=windows GOARCH=amd64 go build -o 'bin/$(BINARY_NAME)-windows-amd64'

.PHONY: clean
clean:
				go clean
				rm -f bin/$(BINARY_NAME)-*
