.PHONY: build test vet lint run clean deadcode check

BUILD_DIR = ./bin
BINARY = $(BUILD_DIR)/djinni

build: $(BUILD_DIR)
	go build -o $(BINARY) ./main.go

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

deadcode:
	deadcode -test ./...

test:
	go test -v ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./... --fix

check: build vet lint test deadcode

run: build
	./$(BINARY)

clean:
	rm -rf $(BUILD_DIR)
	go clean
