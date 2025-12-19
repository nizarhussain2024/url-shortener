.PHONY: build run clean

build:
	go build -o url-shortener ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -f url-shortener





