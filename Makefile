PHONY: build

make:

run:
	bin/kaw

dev:
	go run main.go

build: build_all

build_bin: 
	go build -o bin/kaw main.go
build_client:
	sudo rm -rf build && \
	cd client && \
	npm run build && \
	cp dist/* ../public \
	cd ..

build_all: build_client build_bin