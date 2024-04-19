# Analyzer

An MPEG toy program written in Go. The program can be built with `go >= 1.13` and `make`.

## Quick start

```
$> git clone git@github.com:tony-507/build-tools.git
$> cd analyzers
$> make
```

Built output can be found at `analyzers/build/`. Integration test output can be found at `analyzers/test/output`.

## Architecture

There are several components in the application.

### Plugins

A plugin is specified for a type of job, e.g. ioReader plugin is responsible for reading input, tsdemux demultiplexes a transport stream.

All plugins implement a set of interfaces for handling the data flow.

### Kernel

A kernel coordinates different plugins by calling their interfaces. In this project, tttKernel is the package containing the worker. It exposes a simple API for one to run the application.
```
tttKernel.StartService(params []OverallParams, selectPlugin func(string) IPlugin)
```

The `params` parameter describes how the properties of plugins used, including their properties and how they are connected. It allows one to configure plugins' parameters and construct a workflow among plugins. The `selectPlugin` parameter provides a plugin selector for kernel to select appropriate plugin for each plugin in `params`.

For detail on syntax of the script, refer to documentation in tttKernel.

### Controller

A controller is responsible for determining the parameters passed to the kernel according to the application requirement.

The application requirement is specified by a `.ttt` file.

## Testing

To run test:
```
$> make run-test
```

To generate coverage report:
```
$> make coverage
```
