package cl

// This file provides CGO flags to find OpecnCL libraries and headers.

//#cgo darwin LDFLAGS: -framework OpenCL
//#cgo !darwin LDFLAGS: -lOpenCL
//
////default location:
//#cgo LDFLAGS:-L/usr/lib/x86_64-linux-gnu/
//#cgo CFLAGS: -I/usr/include
//
////Ubuntu 15.04::
//#cgo LDFLAGS:-L/usr/lib/x86_64-linux-gnu/
//#cgo CFLAGS: -I/usr/include
//
////arch linux:
//#cgo LDFLAGS:-L/opt/lib64 -L/opt/lib
//#cgo CFLAGS: -I/opt/include
//
////WINDOWS:
//#cgo windows LDFLAGS:-LC:/Intel/opencl/lib/x64
//#cgo windows CFLAGS: -IC:/Intel/opencl/include
import "C"
