#!/bin/bash

showHelp() {
	echo "Usage: sh build.sh <options> <flags>"
}

init() {
	curDir=$(pwd)
	echo "Current directory: $curDir"

	echo "Begin CI/CD workflow"
	rm -rf build

	mkdir -p build
	cd build
}

build_app() {
	echo -e "build_tsa:\n"
	go build $curDir/cmd/tsa/main.go
	mv main tsa

	echo -e "build_editCap:\n"
	go build $curDir/cmd/editCap/main.go
	mv main editCap
}

build_test() {
	echo -e "build_unitTest:\n"
	go build $curDir/cmd/unitTest/main.go
	mv main unitTest
}

run_test() {
	./unitTest
}

runTest=true

userBuild () {
	# Fail on any error
	set -e

	init

	build_app

	build_test

	run_test
}
