.PHONY: test golden-update

FILE_NAME := remixdb-$(shell go env GOOS)-$(shell go env GOARCH)
ifeq ($(shell go env GOOS),windows)
FILE_NAME := $(FILE_NAME).exe
endif

build:
	go generate ./...
	mkdir -p ./bin
	go build -o ./bin/$(FILE_NAME) ./cmd/remixdb

test:
	go generate ./...
	go run ./internal/cmd/pretestsetup/pretestsetup.go
	go test ./... -cover -race

golden-update:
	go generate ./...
	go run ./internal/cmd/pretestsetup/pretestsetup.go
	GOLDEN_UPDATE=1 go test ./... -cover -race

install:
	make build
	cp ./bin/$(FILE_NAME) /usr/local/bin/remixdb
