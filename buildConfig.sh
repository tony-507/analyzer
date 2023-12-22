#!/bin/bash

buildDir="${MODULE_DIR}/build"
testOutputDir="${MODULE_DIR}/test/output"

init() {
	for d in ${MODULE_DIR}/{build,build/bin,build/test_result,test/output}; do
		rm -rf $d

		mkdir -p $d
	done
}

build_app() {
	cd $MODULE_DIR

	# Need relative to project base directory
	BUILD_TAG=$(date -u '+%Y_%m_%d_%H_%M_%S')
	go build \
		-ldflags "-X github.com/tony-507/analyzers/src/controller.AppVersion=$BUILD_TAG" \
		-o $buildDir/bin/ ./cmd/...

	cd ..
}

run_unit_tests () {
	go install github.com/jstemmer/go-junit-report
	binPath="$(go env GOPATH)/bin/"
	pkgList=$(go list ./... | grep -v /resources | grep -v /logs | grep -v /testUtils | grep -v /cmd) 

	go test -v -cover -coverpkg ./... -coverprofile $MODULE_DIR/build/test_result/testCoverage.txt ./... -timeout 10m 2>&1 | tee /dev/tty | $binPath/go-junit-report > $MODULE_DIR/build/test_result/test_detail.xml
}

generate_coverage_report () {
	echo "Generating code coverage report"
	while read p || [[ -n $p ]]; do
		sed -i'' -e "/${p//\//\\/}/d" $MODULE_DIR/build/test_result/testCoverage.txt
	done <$MODULE_DIR/.testCoverageIgnore
	go tool cover -html=$MODULE_DIR/build/test_result/testCoverage.txt -o $MODULE_DIR/build/test_result/testCoverage.html
}

userBuild () {
	# Fail on any error
	set -e

	init

	build_app

	# Copy demo apps
	mkdir $buildDir/bin/.resources
	cp $MODULE_DIR/test/resources/apps/* $buildDir/bin/.resources/.
}

userTest() {
	cd $MODULE_DIR

	run_unit_tests

	generate_coverage_report

	cd ..
}
