# AvContainer

This module helps demuxing a transport stream. It accepts TS packet one by one and holds the PSI information for demuxing.

## Features

This module aims to support the following features.

* Parsing of TS header and adaptation field
* Parsing of different PSIs and PES header
* [WIP] Calculation of PCR on stream PES
* [WIP] Upon the arrival of updated PSI information, the parsing references to the new information if we are parsing packets later than that PSI

## Technical Detail

This module employs the following desgin.

### Architectural Design

The demuxer is composed of several components: tsDemuxPipe, demuxMonitor, demuxController. They can communicate with each other internally. This implementation allows separation of concerns and makes debugging and development easier.

1. tsDemuxPipe: This is where demultiplexing is done. It receives packet data from demuxer and parse the information.
2. demuxMonitor: This module is a goroutine that monitors demuxer's work. It detects and monitors states of the demuxer, e.g. it can detect if the demuxer gets stuck and prints useful debugging information.
3. demuxController: This is an internal controller that handles internal requests, e.g. raising error and signalling an alarm.

### FAQs

1. Why not using multithreading in demuxPipe?

It is inefficient to use it there. We have a lot of race conditions in demuxPipe, so multithreading makes things complicated. For example, if PSI is updated, PES data needs to wait for its parsing before going on. Another example is the stamping of PCR that requires information from the PCR-carrying stream in advance.