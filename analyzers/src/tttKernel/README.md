# tttKernel

This is the kernel module for the application. It provisions a  script and input arguments to build a graph connecting different plugins.

## Plugin Management

This application does not support dynamic library loading.

To add a new plugin,
* Create a class inheriting the interface `IPlugin`.
* Register the plugin with your own plugin selector function.

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
