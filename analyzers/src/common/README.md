# Common package

This package provides common utilities for the application.

* CmUnit: This is the interface for delivery of data between plugins.
* CmBuf: This is the interface for buffer contained in a unit.
* Plugin: This is the interface for plugins.
* Callback: This file contains functions for plugin communications.

# CmUnit

There are several types of units.

## StatusUnit

This unit allows communication between non-neighbouring plugins. It should not contain any buffer and should be used for plugin's configuration update. A plugin needs to send a request to worker to indicate that it wants to listen to the request, and this should be done during SetParameter.