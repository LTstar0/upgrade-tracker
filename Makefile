.PHONY: deps run build tidy build-linux-x86 package-linux-x86

deps:
	go mod tidy
	go mod download

run:
	go run ./cmd/server/main.go

build:
	go build -o upgrade-tracker-server ./cmd/server/main.go

build-linux-x86:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o upgrade-tracker-server-linux-x86 ./cmd/server/main.go

package-linux-x86:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o upgrade-tracker-server ./cmd/server/main.go
	mkdir -p upgrade-tracker-release
	cp upgrade-tracker-server config.yaml upgrade-tracker-release/
	cp -r frontend sql upgrade-tracker-release/
	tar -czvf upgrade-tracker-release-linux-x86.tar.gz upgrade-tracker-release/
	rm -rf upgrade-tracker-release upgrade-tracker-server

tidy:
	go mod tidy

