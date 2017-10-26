# GoCrack Build Tools

This repository includes docker containers that build the necessary components for a successful gocrack install.

* Current Hashcat Version: v3.6
* Current Go Version: v1.9

**Note**: You should be using Docker 17.X and require NVIDIA Docker plugin to run the worker image

## Building GoCrack Images

    $ make build

Running this will create 4 images:

1. `gocrack/hashcat_shared` builds hashcat as a shared library for Ubuntu 16.04 (Xenial) and exports the build to `dist/hashcat`
1. `gocrack/build` builds gocrack using the hashcat library from `gocrack/hashcat_shared` and exports two binaries to `dist/gocrack`
1. `gocrack/server` will be a Ubuntu 16.04 container with GoCrack server components
1. `gocrack/worker` will be a Ubuntu 16.04 container with the worker components and NVIDIA components needed for OpenCL

## Current Image Sizes

Build Images:

    gocrack/build               633MB
    gocrack/hashcat_shared      468MB

Deployment Images:

    gocrack/worker              194MB
    gocrack/server              198MB