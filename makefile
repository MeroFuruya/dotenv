all: build

build:
	go build -o build/dotenv .

clean:
	rm -rf build

install: build
	mv build/dotenv /usr/local/bin/dotenv

.PHONY: all build clean install