.DEFAULT_GOAL = dev
PROG          = de-swap
SRC           = main.go

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

dev: mod fmt vet build

.PHONY: fmt build vet mod dev linux
