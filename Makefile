.PHONY: build docker build-all

build-all: build-linux-amd64 build-linux-arm64

build-test:
	@go mod tidy
	@rm -f ./build/deeplx_* || true
	@go build -o deeplx .

build-linux-amd64:
	@mkdir -p build
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -ldflags "-s -w" -o build/deeplx_linux_amd64 .

build-linux-arm64:
	@mkdir -p build
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -gcflags "all=-N -l" -ldflags "-s -w" -o build/deeplx_linux_arm64 .

build-mac-arm64:
	@mkdir -p build
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -gcflags "all=-N -l" -ldflags "-s -w" -o build/deeplx_mac_arm64 .

build-win-arm64:
	@mkdir -p build
	@CGO_ENABLED=1 GOOS=windows GOARCH=arm64 go build -gcflags "all=-N -l" -ldflags "-s -w -H windowsgui" -o build/deeplx_win_arm64.exe .


gox-linux:
	gox -gcflags="all=-N -l" -ldflags "-s -w" -osarch="linux/amd64 linux/arm64" -output="build/deeplx_{{.OS}}_{{.Arch}}"

gox-all:
	gox -gcflags="all=-N -l" -ldflags "-s -w" -osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64" -output="build/deeplx_{{.OS}}_{{.Arch}}"

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