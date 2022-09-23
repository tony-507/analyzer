# Controller

Controller module provides a uniform interface to create a service with the worker. When users configure parameters for the service, the parameters are parsed into a JSON object by the user interface backend handler. The resulting JSON is passed to the controller that in turn sets worker parameters using the JSON.