# Go software for automated measurements 

Copyright Jan Kåre Vatne, 2020

This package implements interfaces for different multimeters, oscilloscopes and power supplies.
It is independent of National instruments drivers or other software. The only non-google imports
are the serial library, which is my own fork of github.com/tarm/serial, with some modifications,
and the small logger project on github.com/jkvatne/alog

It is is developed for Windows 10, but all tcp/ip and serial interfaces should work on Linux.
The only exception is the Digilent Analog Discovery 2, which uses a binary library. But this
is also available in Linux, so it should be possible to implement a Linux version.

NB: This is a work in progress, and things may change dramatically when new functions are added.

Instruments support will be extended later. The following are currently supported:

### Multimeters
* Fluke 8845A

### Power supplies
* TTi CPX400
* Korad KD3005
* Manual controlled supply

### Oscilloscopes
* Tektronix TDS2000 series

### Multifunctions instruments
* Digilent Analog Discovery 2

--