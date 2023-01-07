#!/bin/bash

buildDir="${MODULE_DIR}/build"

init() {
	rm -rf $buildDir

	mkdir -p $buildDir
}

build_app() {
	cd $MODULE_DIR

	for d in $MODULE_DIR/cmd/*; do
		app=$(basename $d)
		echo "Building $app"

		go build $d/main.go
		mv $MODULE_DIR/main $buildDir/$app
		echo ""
	done

	cd ..
}

userBuild () {
	# Fail on any error
	set -e

	init

	build_app
}

userTest() {
	cd $MODULE_DIR
	go install github.com/jstemmer/go-junit-report
	binPath="$(go env GOPATH)/bin/"
	pkgList=$(go list ./... | grep -v /resources | grep -v /logs | grep -v /testUtils | grep -v /cmd) 

	go test -v -cover -coverpkg ./... -coverprofile $MODULE_DIR/build/unitTestCoverage.txt ./... -timeout 5s 2>&1 |  $binPath/go-junit-report > $MODULE_DIR/build/test_detail.xml

	echo "Generating code coverage report"
	while read p || [[ -n $p ]]; do
		sed -i'' -e "/${p//\//\\/}/d" $MODULE_DIR/build/unitTestCoverage.txt
	done <$MODULE_DIR/.unitTestCoverageIgnore
	go tool cover -html=$MODULE_DIR/build/unitTestCoverage.txt -o $MODULE_DIR/build/unitTestCoverage.html
	cd ..
}
