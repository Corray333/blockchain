.SILENT:
build:
	go build cmd/main.go
	mv main cmd
run: build
	./cmd/main
test:
	go test ./... -count=1