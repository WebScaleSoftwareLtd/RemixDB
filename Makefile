.PHONY: test golden-update

test:
	go run ./internal/pretestsetup/pretestsetup.go
	go test ./... -cover -race

golden-update:
	go run ./internal/pretestsetup/pretestsetup.go
	GOLDEN_UPDATE=1 go test ./... -cover -race
