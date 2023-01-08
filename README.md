# Analyzer

An MPEG toy program written in Go. The program can be built with Go >= 1.13.

## Quick start

Fetch common build scripts
```
$> git clone git@github.com:tony-507/build-tools.git
$> ln -sf build-tools/build-tools/build-local.sh build-local.sh
```

Run build flow
```
$> ./build-local.sh
```

Built output can be found at `analyzers/build/`. Integration test output can be found at `analyzers/test/output`.

## Architecture

This project uses plugin-worker architecture. A plugin performs a particular job and a worker coordinates between different plugins. In this project, tttKernel is the package containing the worker. It exports a simple API
```
tttKernel.StartApp(script string, input []string)
```
for one to run the application. The script parameter is a `.ttt` file with specified syntax that describes the plugins. It allows one to configure plugins' parameters and construct a workflow among plugins. The input parameter is the list of input arguments that provide further flexibility on using a script.

For detail on syntax of the script, refer to documentation in tttKernel.

## Testing

This projects contains codes for unit tests and integration tests using Go's native testing package. One can check the coverage of the tests with the output html file.