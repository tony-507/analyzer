# Data Handler

This module is responsible for handling PES data. It holds a list of handlers that handle data from respective streams.

## Implementation

This module runs handlers as Goroutines to allow better performance.

It grabs tha data from demuxer's output and directly passes through the unit to next plugin, so it won't affect the high level data flow. This also allows one to disable the use of this module in a simple way.