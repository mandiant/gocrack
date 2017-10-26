package gocat

// #include <stdlib.h>
import "C"
import "unsafe"

var cChar *C.char

// convertArgsToC converts go strings into a **C.char.
// The results of this MUST be free'd
func convertArgsToC(args ...string) (C.int, **C.char) {
	ptrSize := unsafe.Sizeof(cChar)
	ptr := C.malloc(C.size_t(len(args)) * C.size_t(ptrSize))

	for i := 0; i < len(args); i++ {
		element := (**C.char)(unsafe.Pointer(uintptr(ptr) + uintptr(i)*ptrSize))
		*element = C.CString(string(args[i]))
	}
	return C.int(len(args)), (**C.char)(ptr)
}
