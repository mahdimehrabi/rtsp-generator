devtools:
	@echo "Installing devtools"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
lint:
	golangci-lint run --config .golangci.yml

fmt:
	gofumpt -l -w .;gci write ./

build:
	docker build -t buildf .