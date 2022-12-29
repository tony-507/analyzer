#!/bin/bash

buildDir="${MODULE_DIR}/build"

init() {
	rm -rf $buildDir

	mkdir -p $buildDir
}

build_app() {
	cd $MODULE_DIR

	echo -e "build_tsa:\n"
	go build ${MODULE_DIR}/cmd/_tsa_/main.go
	mv $MODULE_DIR/main $buildDir/tsa

	echo -e "build_editCap:\n"
	go build ${MODULE_DIR}/cmd/_editCap_/main.go
	mv $MODULE_DIR/main $buildDir/editCap

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
	go test $(go list ./... | grep -v /resources | grep -v /logs | grep -v /testUtils) -coverprofile $MODULE_DIR/build/unitTestCoverage.txt
	cd ..
}
