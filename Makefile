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
	@docker rmi -f neccen/deeplx-local || true
	@docker build -t neccen/deeplx-local .
	@docker-buildx build --platform linux/arm64 -t neccen/deeplx-local .
	@docker run -itd -p 62155:62155 --name deeplx neccen/deeplx-local