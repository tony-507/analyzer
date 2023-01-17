# AvContainer

This module helps demuxing a transport stream. It accepts TS packet one by one and holds the PSI information for demuxing.

## Features

This module performs:

* Demultiplexing of TS to get the data type of each elementary stream
* PCR interpolation on TS packets
* Validations based on TS header and PES header

## Technical Detail

This module consists of three parts: tsdemux, tsmux and model.

### model

This directory contains utilities to convert between bitstreams and readable data.

### tsdemux

The demuxer is composed of two components: tsDemuxPipe and demuxController. They can communicate with each other internally. This implementation allows separation of concerns and makes debugging and development easier.

1. tsDemuxPipe: This is where demultiplexing is done. It receives packet data from demuxer and parse the information.
3. demuxController: This is an internal controller that handles internal requests, e.g. raising error and signalling an alarm.

### FAQs

1. Why not using multithreading in demuxPipe?

It is inefficient to use it there. We have a lot of race conditions in demuxPipe, so multithreading makes things complicated. For example, if PSI is updated, PES data needs to wait for its parsing before going on. Another example is the stamping of PCR that requires information from the PCR-carrying stream in advance.