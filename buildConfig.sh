#!/bin/bash

buildDir="${MODULE_DIR}/build"

init() {
	rm -rf $buildDir

	mkdir -p $buildDir
}

build_app() {
	cd $MODULE_DIR

	#for d in $MODULE_DIR/cmd/*; do
	#	app=$(basename $d)
	#	echo "Building $app"
#
#		go build $d/main.go
#		mv $MODULE_DIR/main $buildDir/$app
#		echo ""
#	done

	go build $MODULE_DIR/cmd/ttt/main.go
	mv $MODULE_DIR/main $buildDir/ttt

	echo "Copy apps to .resources"
	cp -r $MODULE_DIR/resources/apps $buildDir/.resources

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
	go test $(go list ./... | grep -v /resources | grep -v /logs | grep -v /testUtils | grep -v /cmd) -coverprofile $MODULE_DIR/build/unitTestCoverage.txt
	cd ..
}
