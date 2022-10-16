# TS analyzer

An MPEG toy program written in Go. The program can be built with Go >= 1.13.

## Quick start

To execute a CI/CD build flow, run
```$> sh build.sh```
at project root. The build output can be found at `build/`.

The analyzer can be used to demux TS:
```tsa <input.ts>```
The output can be found in the output directory. It consists of PSI information and packet timing information.
## Overall architecture

This architecture is adapted from Harmonic inc's RMP. The implementation behind is as follows:
1. User configures preferences on user interface
2. The preferences are parsed as different sets of parameters
3. A worker is initialized with the given parameters
4. The worker constructs a operation graph of plugins based on the parameters
5. The worker runs the graph

We can see that there are 4 main components of this architecture:
1. UI
2. Preferences parser (controller)
3. Worker
4. Plugins

This allows separation of concerns and makes debugging and development easier.