# gocat

**This is a work in progress.** *API's will most likely change.*

gocat is a cgo library for interacting with libhashcat. gocat enables you to create purpose-built password cracking tools that leverage the capabilities of [hashcat](https://hashcat.net/hashcat/).

## Installation

gocat requires hashcat [v3.6.0](https://github.com/hashcat/hashcat/releases) or higher to be compiled as a shared library. This can be accomplished by modifying hashcat's `src/Makefile` and setting `SHARED` to `1`

Installing the Go Library:

    go get github.com/fireeye/gocrack/gocat

## Known Issues

* Lack of Windows Support: This won't work on windows as I haven't figured out how to build hashcat on windows
* Memory Leaks: hashcat has several (small) memory leaks that could cause increase of process memory over time

## Contributing

Contributions are welcome via pull requests provided they meet the following criteria:

1. One feature or bug fix per PR
1. Code should be properly formatted (using go fmt)
1. Tests coverage should rarely decrease. All new features should have proper coverage
