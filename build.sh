#!/bin/bash

showHelp() {
	echo "Usage: sh build.sh <options> <flags>"
}

init() {
# curDir=`dirname -- "$(readlink -f -- $0)"`
	curDir=$(pwd)
	echo "Current directory: $curDir"

	echo "Begin CI/CD workflow"
	rm -rf build

	mkdir -p build
	cd build
}

build_app() {
	echo "build_tsa:\n"
	go build $curDir/cmd/tsa/main.go
	mv main tsa

	echo "build_editCap:\n"
	go build $curDir/cmd/editCap/main.go
	mv main editCap
}

build_test() {
	echo "build_unitTest:\n"
	go build $curDir/cmd/unitTest/main.go
	mv main unitTest
}

run_test() {
	./unitTest
}

# Fail on any error
set -e

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

			if $runTest;then
				build_test
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
