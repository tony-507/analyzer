#!/bin/bash

showHelp() {
	echo "Usage: sh build.sh <options> <flags>"
}

init() {
	curDir=`dirname -- "$(readlink -f -- $0)"`
	echo "Current directory: $curDir"

	echo "Begin CI/CD workflow"
	rm -rf build
	mkdir -p build/resources
	cd build
}

build_app() {
	echo "Building tsa\n"
	# Executable
	go build $curDir/cmd/tsa/main.go
	mv main tsa
}

build_test() {
	echo "Building unitTest\n"
	go build $curDir/cmd/unitTest/main.go
	mv main unitTest
}

run_test() {
	echo "Start running tests\n"
	./unitTest
}

runTest=true

if [ "$#" -gt 0 ];then
	case $1 in
		"build")
			case $2 in
				"-x")
					case $3 in
						"test")
							runTest=false
							;;
						"*")
							showHelp
							;;
						esac
					;;

				"*")
					showHelp
					;;
			esac

			init
			build_app
			build_test

			if $runTest;then
				run_test
			fi
			;;

		"test")
			cd build
			run_test
			;;
		
		"help")
			showHelp
			;;
		"*")
			showHelp
			Exit 1
			;;
	esac
	echo "Done"
else
	showHelp
fi