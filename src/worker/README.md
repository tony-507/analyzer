# Worker Module

This module contains source code for the worker. A worker builds a operation graph from parameters and runs a service based on the graph. It coordinates several plugins to provide the service.

## Working Principle

In principle, a worker starts all plugins, then keeps sending dummy input to plugins that handle inputs. Those plugins would then send POST requests to worker for further action.

## Notes

The graph for the worker employs pointers for resource saving and efficient callback, so one needs to be careful of dangling pointers when using the worker.