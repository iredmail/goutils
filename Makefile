# First command is the default target running `make` without any argument.

# 如需使用国内镜像，请运行：
# `go env -w  GOPROXY=https://goproxy.io,direct`

#GO = ~/.go/bin/go1.15.6
#GO = ~/.go/go1.16/bin/go
GO = CGO_ENABLED=0 go
#GO = go

SPIDER_VERSION = v1.0.0

LDFLAGS=-s -w
TAGS_PROD=-tags prod

# Go 1.18+
BUILD_ARGS=-buildvcs=false -trimpath -ldflags="${LDFLAGS}"

# Build for development.
GO_BUILD_DEV = ${GO} build ${BUILD_ARGS}
GO_BUILD_PROD = ${GO} build ${BUILD_ARGS} ${TAGS_PROD}

# Builds for production.
GO_BUILD_LINUX_ARM6 = GOOS=linux GOARCH=arm GOARMVERSION=6 ${GO_BUILD_PROD}
GO_BUILD_LINUX_ARM7 = GOOS=linux GOARCH=arm GOARMVERSION=7 ${GO_BUILD_PROD}
GO_BUILD_LINUX_ARM64 = GOOS=linux GOARCH=arm GOARMVERSION=arm64 ${GO_BUILD_PROD}
GO_BUILD_LINUX_AMD64 = GOOS=linux GOARCH=amd64 ${GO_BUILD_PROD}
GO_BUILD_OPENBSD_AMD64= GOOS=openbsd GOARCH=amd64 ${GO_BUILD_PROD}
GO_BUILD_OPENBSD_ARM6 = GOOS=openbsd GOARCH=arm GOARMVERSION=6 ${GO_BUILD_PROD}
GO_BUILD_OPENBSD_ARM64= GOOS=openbsd GOARCH=arm64 ${GO_BUILD_PROD}
GO_BUILD_FREEBSD_AMD64= GOOS=freebsd GOARCH=amd64 ${GO_BUILD_PROD}
GO_BUILD_WIN_AMD64= GOOS=windows GOARCH=amd64 ${GO_BUILD_PROD}

test:
	go test -ldflags="${LDFLAGS}" -cover -count=1 -parallel=10 ./...

# Test with verbose output.
testv:
	go test -ldflags="${LDFLAGS}" -v -cover -count=1 -parallel=10 ./...

lichen:
	lichen --config=samples-dev/lichen.yml bin/*

clean:
	rm -rf bin data spider-*tar
