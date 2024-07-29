BUILD_FOLDER  ?= bin
BINARY_NAME   ?= bitcoin-handshake

.PHONY: build
build:
	@echo "Building the application..."
	@go build -o $(BUILD_FOLDER)/$(BINARY_NAME) 

.PHONY: test
test:
	@echo "Running tests..."
	@go clean -testcache && go test ./... -cover

.PHONY: run
run: build
	@echo "Running the application..."
	@./$(BUILD_FOLDER)/$(BINARY_NAME)

.PHONY: clean	
clean:
	@echo "Cleaning up..."
	@go clean
	@rm -rf $(BUILD_FOLDER)

