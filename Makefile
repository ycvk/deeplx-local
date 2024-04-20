.PHONY: build docker

build:
	@go mod tidy
	@rm -f deeplx || true
	@go build -o deeplx .

docker:
	@go mod tidy
	@rm -f deeplx || true
	@GOOS=linux GOARCH=arm64 go build -o deeplx .
	@docker stop deeplx || true
	@docker rm deeplx || true
	@docker rmi -f ycvk:deeplx || true
	@docker build -t ycvk:deeplx .
	@docker-buildx build --platform linux/arm64 -t ycvk:deeplx .
	@docker run -d -p 62155:62155 --name deeplx ycvk:deeplx