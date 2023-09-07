#!/bin/bash

buildDir="${MODULE_DIR}/build"
testOutputDir="${MODULE_DIR}/test/output"

init() {
	for d in ${MODULE_DIR}/{build,test/output}; do
		rm -rf $d

		mkdir -p $d
	done
}

build_app() {
	cd $MODULE_DIR

	# Need relative to project base directory
	go build -o $buildDir/ ./cmd/...

	cd ..
}

run_unit_tests () {
	go install github.com/jstemmer/go-junit-report
	binPath="$(go env GOPATH)/bin/"
	pkgList=$(go list ./... | grep -v /resources | grep -v /logs | grep -v /testUtils | grep -v /cmd) 

	go test -v -cover -coverpkg ./... -coverprofile $MODULE_DIR/build/testCoverage.txt ./... -timeout 60s 2>&1 |  $binPath/go-junit-report > $MODULE_DIR/build/test_detail.xml
}

generate_coverage_report () {
	echo "Generating code coverage report"
	while read p || [[ -n $p ]]; do
		sed -i'' -e "/${p//\//\\/}/d" $MODULE_DIR/build/testCoverage.txt
	done <$MODULE_DIR/.testCoverageIgnore
	go tool cover -html=$MODULE_DIR/build/testCoverage.txt -o $MODULE_DIR/build/testCoverage.html
}

userBuild () {
	# Fail on any error
	set -e

	init

	build_app

	# Copy demo apps
	mkdir $buildDir/.resources
	cp $MODULE_DIR/test/resources/apps/* $buildDir/.resources/.
}

userTest() {
	cd $MODULE_DIR

	run_unit_tests

	generate_coverage_report

	cd ..
}
