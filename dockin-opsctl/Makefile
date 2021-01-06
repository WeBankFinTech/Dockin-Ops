BINARY="dockin-opsctl"
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
	@GOOS=${OS} GOARCH=${ARCH} go build -o ${BINARY} ./cmd  
	# -ldflags=-X dockin-opsctl/internal/version.BuildDate=#date -X dockin-opsctl/internal/version.CommitId=#commitId -X dockin-opsctl/internal/version.Version=#version

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

.PHONY: default fmt fmt-check install vet clean

