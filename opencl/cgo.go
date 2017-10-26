package opencl

/*
#cgo !darwin LDFLAGS: -lOpenCL
#cgo darwin LDFLAGS: -framework OpenCL
#cgo linux CFLAGS: -I${SRCDIR}/deps/OpenCL-Headers
*/
import "C"
