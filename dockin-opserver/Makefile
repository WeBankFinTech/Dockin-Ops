BINARY="dockin-opserver"
VERSION=1.0.0
BUILD=`date +%FT%T%z`

PACKAGES=`go list ./... | grep -v /vendor/`
VETPACKAGES=`go list ./... | grep -v /vendor/ | grep -v /examples/`
GOFILES=`find . -name "*.go" -type f -not -path "./vendor/*"`

OS=linux
ARCH=amd64

default:
	@go mod tidy
	@go mod vendor
	@mkdir build
	@cd build && GOOS=${OS} GOARCH=${ARCH} go build -o ${BINARY} ../cmd


list:
	@echo ${PACKAGES}
	@echo ${VETPACKAGES}
	@echo ${GOFILES}

fmt:
	@gofmt -s -w ${GOFILES}

fmt-check:
	@diff=?(gofmt -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

# test:
#         @go test -cpu=1,2,4 -v -tags integration ./...

vet:
	@go vet $(VETPACKAGES)

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

package: default
	@mkdir -p release/${BINARY}
	@mkdir -p release/${BINARY}/apps
	@mkdir -p release/${BINARY}/bin
	@cp build/${BINARY} release/${BINARY}/apps
	@cp scripts/* release/${BINARY}/bin
	@cp -r configs release/${BINARY}/
	@chmod +rx release/${BINARY}/*
	@cd release && tar -czf "${BINARY}_${VERSION}.tar.gz" ./${BINARY}

.PHONY: default package fmt fmt-check install vet clean
