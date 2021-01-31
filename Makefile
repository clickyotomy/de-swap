PROG          = de-swap
SRC           = main.go

all: mod fmt vet build

fmt:
	go fmt ./...

build:
	go build -o ${PROG} ./...

vet:
	go vet ./...

mod:
	go mod download
	go mod tidy
	go mod verify

.PHONY: fmt build vet mod all
