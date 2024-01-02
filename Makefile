PHONY: build

build: build_all

build_bin: 
	go build -o bin/kaw main.go
build_client:
	cd client && \
	npm run build-web && \
	cp -r web-build/* ../public && \
	cd ..

build_all: build_client build_bin

run:
	bin/kaw

dev:
	go run main.go
