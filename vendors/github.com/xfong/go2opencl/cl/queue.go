package cl

/*
#include "./opencl.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

//////////////// Basic Types ////////////////
type CommandQueueProperty int

const (
	CommandQueueOutOfOrderExecModeEnable CommandQueueProperty = C.CL_QUEUE_OUT_OF_ORDER_EXEC_MODE_ENABLE
	CommandQueueProfilingEnable          CommandQueueProperty = C.CL_QUEUE_PROFILING_ENABLE
)

type CommandQueueInfo int

const (
	CommandQueueContext        CommandQueueInfo = C.CL_QUEUE_CONTEXT
	CommandQueueDevice         CommandQueueInfo = C.CL_QUEUE_DEVICE
	CommandQueueReferenceCount CommandQueueInfo = C.CL_QUEUE_REFERENCE_COUNT
	CommandQueueProperties     CommandQueueInfo = C.CL_QUEUE_PROPERTIES
)

//////////////// Abstract Types ////////////////
type CommandQueue struct {
	clQueue C.cl_command_queue
	device  *Device
}

//////////////// Golang Types ////////////////
type CLCommandQueueProperties C.cl_command_queue_properties

//////////////// Basic Functions ////////////////
func retainCommandQueue(q *CommandQueue) {
	if q.clQueue != nil {
		C.clRetainCommandQueue(q.clQueue)
	}
}

func releaseCommandQueue(q *CommandQueue) {
	if q.clQueue != nil {
		C.clReleaseCommandQueue(q.clQueue)
		q.clQueue = nil
	}
}

//////////////// Abstract Functions ////////////////
// Call clRetainCommandQueue on the CommandQueue.
func (q *CommandQueue) Retain() {
	retainCommandQueue(q)
}

// Call clReleaseCommandQueue on the CommandQueue. Using the CommandQueue after Release will cause a panick.
func (q *CommandQueue) Release() {
	releaseCommandQueue(q)
}

// Blocks until all previously queued OpenCL commands in a command-queue are issued to the associated device and have completed.
func (q *CommandQueue) Finish() error {
	return toError(C.clFinish(q.clQueue))
}

// Issues all previously queued OpenCL commands in a command-queue to the device associated with the command-queue.
func (q *CommandQueue) Flush() error {
	return toError(C.clFlush(q.clQueue))
}

func (ctx *Context) CreateCommandQueue(device *Device, properties CommandQueueProperty) (*CommandQueue, error) {
	var err C.cl_int
	clQueue := C.clCreateCommandQueue(ctx.clContext, device.id, C.cl_command_queue_properties(properties), &err)
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	if clQueue == nil {
		return nil, ErrUnknown
	}
	commandQueue := &CommandQueue{clQueue: clQueue, device: device}
	runtime.SetFinalizer(commandQueue, releaseCommandQueue)
	return commandQueue, nil
}

func (q *CommandQueue) GetQueueContext() (*Context, error) {
	if q.clQueue != nil {
		var outContext C.cl_context
		err := C.clGetCommandQueueInfo(q.clQueue, C.CL_QUEUE_CONTEXT, C.size_t(unsafe.Sizeof(outContext)), unsafe.Pointer(&outContext), nil)
		return &Context{clContext: outContext, devices: nil}, toError(err)
	}
	return nil, toError(C.CL_INVALID_COMMAND_QUEUE)
}

func (q *CommandQueue) GetQueueDevice() (*Device, error) {
	if q.clQueue != nil {
		var outDevice C.cl_device_id
		err := C.clGetCommandQueueInfo(q.clQueue, C.CL_QUEUE_DEVICE, C.size_t(unsafe.Sizeof(outDevice)), unsafe.Pointer(&outDevice), nil)
		return &Device{id: outDevice}, toError(err)
	}
	return nil, toError(C.CL_INVALID_COMMAND_QUEUE)
}

func (q *CommandQueue) GetQueueReferenceCount() (CLUint, error) {
	if q.clQueue != nil {
		var outCount C.cl_uint
		err := C.clGetCommandQueueInfo(q.clQueue, C.CL_QUEUE_REFERENCE_COUNT, C.size_t(unsafe.Sizeof(outCount)), unsafe.Pointer(&outCount), nil)
		return CLUint(outCount), toError(err)
	}
	return 0, toError(C.CL_INVALID_COMMAND_QUEUE)
}

func (q *CommandQueue) GetQueueProperties() (CommandQueueProperty, error) {
	if q.clQueue != nil {
		var outVar CommandQueueProperty
		err := C.clGetCommandQueueInfo(q.clQueue, C.CL_QUEUE_PROPERTIES, C.size_t(unsafe.Sizeof(outVar)), unsafe.Pointer(&outVar), nil)
		return outVar, toError(err)
	}
	return 0, toError(C.CL_INVALID_COMMAND_QUEUE)
}
