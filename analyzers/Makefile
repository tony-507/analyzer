SHELL := /bin/bash

default: build
.PHONY: build

clean:
	rm -rf ./{build,build/bin,build/test_result,test/output}
	mkdir -p ./{build,build/bin,build/test_result,test/output}

build:
	go build \
		-ldflags "-X github.com/tony-507/analyzers/src/controller.AppVersion=$$(date -u '+%Y_%m_%d_%H_%M_%S')" \
		-o ./build/bin/ ./cmd/...

run-test: build
	go install github.com/jstemmer/go-junit-report
	binPath="$$(go env GOPATH)/bin/" && \
	pkgList=$$(go list ./... | grep -v /resources | grep -v /logs | grep -v /testUtils | grep -v /cmd) && \
	go test -v -cover -coverpkg ./... -coverprofile ./build/test_result/testCoverage.txt \
		./... -timeout 10m 2>&1 | \
		tee /dev/tty | \
		$$binPath/go-junit-report > ./build/test_result/test_detail.xml

coverage: run-test
	while read p || [[ -n $$p ]]; do \
		sed -i'' -e "/$${p//\//\\/}/d" ./build/test_result/testCoverage.txt; \
	done <./.testCoverageIgnore
	go tool cover -html=./build/test_result/testCoverage.txt -o ./build/test_result/testCoverage.html
