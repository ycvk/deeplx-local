.PHONY: build docker

build:
	@go mod tidy
	@rm -f deeplx || true
	@GOOS=darwin GOARCH=arm64 go build -o deeplx .

docker:
	@go mod tidy
	@rm -f deeplx || true
	@GOOS=linux GOARCH=arm64 go build -o deeplx .
	@docker stop deeplx || true
	@docker rm deeplx || true
	@docker rmi -f deeplx:v0.0.1 || true
	@docker build -t deeplx:v0.0.1 .
	@docker-buildx build --platform linux/arm64 -t deeplx:v0.0.1 .
	@docker run -d -p 62155:62155 --name deeplx deeplx:v0.0.1