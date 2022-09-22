#!/bin/bash

curDir=`dirname -- "$(readlink -f -- $0)"`
echo "Current directory: $curDir"

echo "Begin CI/CD workflow"
rm -rf build
mkdir -p build/resources
cd build

build_opt=0 # 0: full build, 1: build without test

if (($1)); then
	build_opt=1
fi

echo "Building tsa\n"
# Executable
go build $curDir/cmd/tsa/main.go
mv main tsa
# Resources for app
cp $curDir/src/resources/app.json $curDir/build/resources/.

echo "Building unitTest\n"
go build $curDir/cmd/unitTest/main.go
mv main unitTest

echo "Start running tests\n"
./unitTest