# TS analyzer

An MPEG toy program written in Go.

## Overall architecture

This architecture is adapted from Harmonic inc's RMP. Initially there is also a controller layer, but I think it makes things too complicated here, so it is omitted for now.

--------------------------------------------------
|                                                |
|           User interface (CLI, API,...)        |
|                                                |
--------------------------------------------------
                        |
                        | (parameters passed)
                        v
--------------------------------------------------
|                                                |
|                     Worker                     |
|                                                |
--------------------------------------------------
                        |
                        | (calls plugin interfaces)
                        v
--------------------------------------------------
|                                                |
|                     Plugins                    |
|                                                |
--------------------------------------------------
