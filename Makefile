# Detect OS: on Windows, $(OS) is set to "Windows_NT"
ifeq ($(OS),Windows_NT)
    BINARY_NAME := screensaver.exe
    RM          := del /Q
else
    BINARY_NAME := screensaver
    RM          := rm -f
endif

CMD := ./cmd/screensaver

.PHONY: all build install test clean

all: build

## build: compile the binary for the current platform
build:
	go build -o $(BINARY_NAME) $(CMD)

## install: install via `go install` (places binary in $(GOPATH)/bin)
install:
	go install $(CMD)

## test: run all unit tests
test:
	go test ./...

## clean: remove build artefacts
clean:
	$(RM) screensaver screensaver.exe
