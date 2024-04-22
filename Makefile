.PHONY: build docker

build:
	@go mod tidy
	@rm -f ./build/deeplx_* || true
	@go build -o deeplx .

gox-linux:
	gox -gcflags="all=-N -l" -ldflags "-s -w -X 'main.GO_VERSION=$(go version)'" -osarch="linux/amd64 linux/arm64" -output="build/deeplx_{{.OS}}_{{.Arch}}"

gox-all:
	gox -gcflags="all=-N -l" -ldflags "-s -w -X 'main.GO_VERSION=$(go version)'" -osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64" -output="build/deeplx_{{.OS}}_{{.Arch}}"

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