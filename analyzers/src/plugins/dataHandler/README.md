# Data Handler

This module is responsible for handling stream data. It holds a list of handlers that handle data from respective streams.

## Implementation

This module runs handlers as Goroutines to allow better performance.

It receives data from demuxer's output and perform further parsing.