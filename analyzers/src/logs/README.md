# Logs module

This module implements a logger.

## Expected System Behaviour

To create a logger:
```
logger := logs.GetLogger("identifier")
```

To set logger configuration programatically:
```
logs.SetProperty("propertyName", "propertyValue")
```
Configuration details can be found below.

To write a log:
```
logger.Log(logs.LEVEL, "message")
```

## Architecture

The logs module stores configuration globally. The identifier identifies where the log comes from as a parameter in display form.