# TS analyzer

An MPEG toy program written in Go.

## Overall architecture

This architecture is adapted from Harmonic inc's RMP.

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
|                   Controller                   |
|                                                |
--------------------------------------------------
                        |
                        | (parameters formatted)
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
