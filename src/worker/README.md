# Worker Module

This module contains source code for the worker. A worker builds a operation graph from parameters and runs a service based on the graph. It coordinates several plugins to provide the service.

## Working Principle

In principle, a worker starts all plugins, then keeps sending dummy input to plugins that handle inputs. Plugins can then send requests to the worker to continue the workflow. Common request types are:
* `DELIVER_REQUEST`: Ask worker to deliver a unit to the plugin
* `FETCH_REQUEST`: Ask worker to fetch a unit from the plugin
* `EOS_REQUEST`: Inform worker that the plugin can be stopped

Each plugin has an interface in this module to prevent cyclic import, so one can have an overview of what plugins are available from this module.