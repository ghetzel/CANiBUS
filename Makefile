.PHONY: build ui

all: fmt deps build

fmt:
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	esc -o webserver/static.go -pkg webserver -prefix www www
	#goimports -w .

deps:
	go get ./...

build: fmt
	go build -i -o bin/canibusd canibusd/main.go