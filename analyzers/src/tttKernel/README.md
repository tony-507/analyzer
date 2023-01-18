# tttKernel

This is a kernel module for ttt application. It provisions a ttt script and input arguments to build a work graph.

## APIs

tttKernel exports only a single API:
```
tttKernel.StartApp(script string, input []string)
```
One just needs to provide the ttt script as well as the input arguments in order to run the application.

## Plugin Management

This module does not support dynamic library loading. If one would like to add a customised plugin, please add an entry in graphNode.go to allow kernel to get the plugin.

The plugin needs to have the type common.Plugin. This can be done easily via a provided function in common module:
```
CreatePlugin(
	name string,
	isRoot bool,
	setCallback func(RequestHandler),
	setParameter func(string),
	setResource func(*ResourceLoader),
	startSequence func(),
	deliverUnit func(CmUnit),
	deliverStatus func(CmUnit),
	fetchUnit func() CmUnit,
	endSequence func(),
)
```

## ttt Scripting Syntax

This section describes syntax for ttt script.

tttKernel reads the script line by line. Each line is separated by a semi-colon.

### Description

One can add a one-line comment on the first line of the script as a description. Comment can be inserted with the syntax
```
// Comment here
```

### Variable Assignment
```
x = #Dummy_1;
```
A RHS value beginning with a # means a plugin. The kernel would attempt to find a plugin called "Dummy" and assign it with the name "Dummy_1". The variable x would store this plugin.

```
y = $f | hi;
```
A RHS value beginning with $ means to read the value from input arguments. Here y would take the value of the "-f" argument. If this argument is not supplied, y would take the default value after |, i.e. "hi".
```
x.a = hi;
```
A dot syntax means to add an attribute. Here the attribute a of x has been assigned a string value "hi".

### Control Flow
```
link(x, y);
```
A link statement means to connect two plugins together. Here output of x would be sent to y.
```
alias(file, f);
```
This maps "-file" argument to "-f" argument.
```
if <condition>;
    <do something>;
end;
```
This is the syntax of an if statement. Currently condition could only be checking nullness of a variable.