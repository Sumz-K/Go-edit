DIR := ./cmd/go-edit
APP := go-edit
.PHONY: run


run:
	@cd $(DIR) && go run . 

runfile:
	@cd $(DIR) && go run main.go ./targetfile.txt

.PHONY: runfile

.PHONY: build

build:
	@cd $(DIR) && go build -o $(APP)

