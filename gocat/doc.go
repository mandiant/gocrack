/*
Package gocat is a cgo interface around libhashcat. It's main purpose is to abstract hashcat and allow you to build tools in Go that leverage the hashcat engine.

gocat should be used with libhashcat v3.6.0 as previous versions have known memory leaks and could affect long running processes running multiple cracking tasks.

*/
package gocat
