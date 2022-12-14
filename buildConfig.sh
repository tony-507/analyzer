#!/bin/bash

buildDir="${MODULE_DIR}/build"

init() {
	rm -rf $buildDir

	mkdir -p $buildDir
}

build_app() {
	cd $MODULE_DIR

	echo -e "build_tsa:\n"
	go build ${MODULE_DIR}/cmd/tsa/main.go
	mv $MODULE_DIR/main $buildDir/tsa

	echo -e "build_editCap:\n"
	go build ${MODULE_DIR}/cmd/editCap/main.go
	mv $MODULE_DIR/main $buildDir/editCap

	cd ..
}

build_test() {
	cd $MODULE_DIR

	echo -e "build_unitTest:\n"
	go build ${MODULE_DIR}/cmd/unitTest/main.go
	mv $MODULE_DIR/main $buildDir/unitTest

	cd ..
}

userBuild () {
	# Fail on any error
	set -e

	init

	build_app
}

userTest() {
	build_test
	exec ${MODULE_DIR}/build/unitTest
}
