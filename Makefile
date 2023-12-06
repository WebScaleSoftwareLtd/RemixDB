.PHONY: test golden-update

test:
	go run ./internal/cmd/pretestsetup/pretestsetup.go
	go test ./... -cover -race

golden-update:
	go run ./internal/cmd/pretestsetup/pretestsetup.go
	GOLDEN_UPDATE=1 go test ./... -cover -race
