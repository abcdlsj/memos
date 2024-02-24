.PHONY: build

build:
	go generate
	go build

.PHONY: clean
clean:
	go clean

.PHONY: test
test: build
	./memos
	
