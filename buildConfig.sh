#!/bin/bash

showHelp() {
	echo "Usage: sh build.sh <options> <flags>"
}

init() {
	cd ${MODULE_DIR}
	rm -rf build

	mkdir -p build
	cd build
}

build_app() {
	echo -e "build_tsa:\n"
	go build ${MODULE_DIR}/cmd/tsa/main.go
	mv main tsa

	echo -e "build_editCap:\n"
	go build ${MODULE_DIR}/cmd/editCap/main.go
	mv main editCap
}

build_test() {
	echo -e "build_unitTest:\n"
	go build ${MODULE_DIR}/cmd/unitTest/main.go
	mv main unitTest
}

userBuild () {
	# Fail on any error
	set -e

	init

	build_app
}

userTest() {
	build_test

	./unitTest
}