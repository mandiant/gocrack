# opencl

This is based off [go-opencl](https://github.com/samuel/go-opencl) with the following changes:

* Lots of code removed to support only what this project needs
* Removed call to `panic` whenever a function that should never fail, fails. It should be left up to the end user to decide what to do
* Changed how string allocation's occur in the `getInfoString` methods
* OS X Support

## Why?

Hashcat includes code to list OpenCL devices but requires one to fully initialize a hashcat context & session to safely pull the device listing back.