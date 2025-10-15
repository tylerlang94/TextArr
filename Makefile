CMD_DIR=./cmd/textarr
BINARY=textarr

.PHONY: run build clean

run:
	@echo "Running $(BINARY)"
	go mod tidy
	go run $(CMD_DIR)/main.go

build:
	@echo "Building @(BINARY)"
	go mod tidy
	cd $(CMD_DIR) && go build -o ../../bin/$(BINARY)	

clean:
	rm -rf bin/ 
