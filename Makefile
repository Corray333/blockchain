.SILENT:
build:
	go build cmd/main.go
	mv main cmd
run: build
	cd cmd
	./main
test:
	go test ./... -count=1