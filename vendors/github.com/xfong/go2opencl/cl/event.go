package cl

/*
#include "./opencl.h"

extern void go_set_event_callback(cl_event event, cl_int execution_status, void *user_args);
static void CL_CALLBACK c_set_event_callback(cl_event event, cl_int execution_status, void *user_args) {
        go_set_event_callback((cl_event) event, (cl_int) execution_status, (void *)user_args);
}
static cl_int CLSetEventCallback(      cl_event		event,
				       cl_int		callback_type,
                                       void *		user_args) {
	return clSetEventCallback(event, callback_type, c_set_event_callback, user_args);
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// ////////////// Basic Types ///////////////
type ProfilingInfo int

const (
	// A 64-bit value that describes the current device time counter in
	// nanoseconds when the command identified by event is enqueued in
	// a command-queue by the host.
	ProfilingInfoCommandQueued ProfilingInfo = C.CL_PROFILING_COMMAND_QUEUED

	// A 64-bit value that describes the current device time counter in
	// nanoseconds when the command identified by event that has been
	// enqueued is submitted by the host to the device associated with the command-queue.
	ProfilingInfoCommandSubmit ProfilingInfo = C.CL_PROFILING_COMMAND_SUBMIT

	// A 64-bit value that describes the current device time counter in
	// nanoseconds when the command identified by event starts execution on the device.
	ProfilingInfoCommandStart ProfilingInfo = C.CL_PROFILING_COMMAND_START

	// A 64-bit value that describes the current device time counter in
	// nanoseconds when the command identified by event has finished
	// execution on the device.
	ProfilingInfoCommandEnd ProfilingInfo = C.CL_PROFILING_COMMAND_END
)

// ////////////// Abstract Types ///////////////
type Event struct {
	clEvent C.cl_event
}

// //////////////// Supporting Types ////////////////
type CL_go_set_event_callback func(event C.cl_event, callback_status C.cl_int, user_data unsafe.Pointer)

var go_set_event_callback_func map[unsafe.Pointer]CL_go_set_event_callback

// ////////////// Basic Functions ///////////////
//
//export go_set_event_callback
func go_set_event_callback(event C.cl_event, callback_status C.cl_int, user_data unsafe.Pointer) {
	var c_user_data []unsafe.Pointer
	c_user_data = *(*[]unsafe.Pointer)(user_data)
	go_set_event_callback_func[c_user_data[1]](event, callback_status, c_user_data[0])
}

func releaseEvent(ev *Event) {
	if ev.clEvent != nil {
		C.clReleaseEvent(ev.clEvent)
		ev.clEvent = nil
	}
}

func retainEvent(ev *Event) {
	if ev.clEvent != nil {
		C.clRetainEvent(ev.clEvent)
	}
}

// Waits on the host thread for commands identified by event objects in
// events to complete. A command is considered complete if its execution
// status is CL_COMPLETE or a negative value. The events specified in
// event_list act as synchronization points.
//
// If the cl_khr_gl_event extension is enabled, event objects can also be
// used to reflect the status of an OpenGL sync object. The sync object
// in turn refers to a fence command executing in an OpenGL command
// stream. This provides another method of coordinating sharing of buffers
// and images between OpenGL and OpenCL.
func WaitForEvents(events []*Event) error {
	return toError(C.clWaitForEvents(C.cl_uint(len(events)), eventListPtr(events)))
}

func newEvent(clEvent C.cl_event) *Event {
	ev := &Event{clEvent: clEvent}
	runtime.SetFinalizer(ev, releaseEvent)
	return ev
}

func eventListPtr(el []*Event) *C.cl_event {
	if el == nil {
		return nil
	}
	elist := make([]C.cl_event, len(el))
	for i, e := range el {
		elist[i] = e.clEvent
	}
	return (*C.cl_event)(&elist[0])
}

// ////////////// Abstract Functions ///////////////
func (e *Event) Release() {
	releaseEvent(e)
}

func (e *Event) Retain() {
	retainEvent(e)
}

func (e *Event) GetEventProfilingInfo(paramName ProfilingInfo) (int64, error) {
	if e.clEvent != nil {
		var paramValue C.cl_ulong
		if err := C.clGetEventProfilingInfo(e.clEvent, C.cl_profiling_info(paramName), C.size_t(unsafe.Sizeof(paramValue)), unsafe.Pointer(&paramValue), nil); err != C.CL_SUCCESS {
			return 0, toError(err)
		}
		return int64(paramValue), nil
	}
	return int64(-1), toError(C.CL_INVALID_EVENT)
}

func (e *Event) GetCommandQueue() (*CommandQueue, error) {
	if e.clEvent != nil {
		var outQueue C.cl_command_queue
		err := C.clGetEventInfo(e.clEvent, C.CL_EVENT_COMMAND_QUEUE, C.size_t(unsafe.Sizeof(outQueue)), unsafe.Pointer(&outQueue), nil)
		return &CommandQueue{clQueue: outQueue, device: nil}, toError(err)
	}
	return nil, toError(C.CL_INVALID_EVENT)
}

func (e *Event) GetContext() (*Context, error) {
	if e.clEvent != nil {
		var outContext C.cl_context
		err := C.clGetEventInfo(e.clEvent, C.CL_EVENT_CONTEXT, C.size_t(unsafe.Sizeof(outContext)), unsafe.Pointer(&outContext), nil)
		return &Context{clContext: outContext, devices: nil}, toError(err)
	}
	return nil, toError(C.CL_INVALID_EVENT)
}

func (e *Event) GetCommandType() (CommandType, error) {
	if e.clEvent != nil {
		var status C.cl_command_type
		var err C.cl_int
		err = C.clGetEventInfo(e.clEvent, C.CL_EVENT_COMMAND_TYPE, C.size_t(unsafe.Sizeof(status)), unsafe.Pointer(&status), nil)
		if gl_sharing_ext {
			check, gl_command_type := getGlCommandType(status)
			if check {
				return gl_command_type, toError(err)
			}
		}
		if dx9_sharing_ext {
			check, dx9_command_type := getDX9CommandType(status)
			if check {
				return dx9_command_type, toError(err)
			}
		}
		if d3d10_sharing_ext {
			check, d3d10_command_type := getD3D10CommandType(status)
			if check {
				return d3d10_command_type, toError(err)
			}
		}
		if d3d11_sharing_ext {
			check, d3d11_command_type := getD3D11CommandType(status)
			if check {
				return d3d11_command_type, toError(err)
			}
		}
		switch status {
		case C.CL_COMMAND_NDRANGE_KERNEL:
			return CommandNDRangeKernel, toError(err)
		case C.CL_COMMAND_TASK:
			return CommandTask, toError(err)
		case C.CL_COMMAND_NATIVE_KERNEL:
			return CommandNativeKernel, toError(err)
		case C.CL_COMMAND_READ_BUFFER:
			return CommandReadBuffer, toError(err)
		case C.CL_COMMAND_WRITE_BUFFER:
			return CommandWriteBuffer, toError(err)
		case C.CL_COMMAND_COPY_BUFFER:
			return CommandCopyBuffer, toError(err)
		case C.CL_COMMAND_READ_IMAGE:
			return CommandReadImage, toError(err)
		case C.CL_COMMAND_WRITE_IMAGE:
			return CommandWriteImage, toError(err)
		case C.CL_COMMAND_COPY_IMAGE:
			return CommandCopyImage, toError(err)
		case C.CL_COMMAND_COPY_BUFFER_TO_IMAGE:
			return CommandCopyBufferToImage, toError(err)
		case C.CL_COMMAND_COPY_IMAGE_TO_BUFFER:
			return CommandCopyImageToBuffer, toError(err)
		case C.CL_COMMAND_MAP_BUFFER:
			return CommandMapBuffer, toError(err)
		case C.CL_COMMAND_MAP_IMAGE:
			return CommandMapImage, toError(err)
		case C.CL_COMMAND_UNMAP_MEM_OBJECT:
			return CommandUnmapMemObject, toError(err)
		case C.CL_COMMAND_MARKER:
			return CommandMarker, toError(err)
		default:
			return -1, toError(err)
		}
	}
	return -1, toError(C.CL_INVALID_EVENT)
}

func (e *Event) GetStatus() (CommandExecStatus, error) {
	if e.clEvent != nil {
		var status C.cl_int
		err := C.clGetEventInfo(e.clEvent, C.CL_EVENT_COMMAND_EXECUTION_STATUS, C.size_t(unsafe.Sizeof(status)), unsafe.Pointer(&status), nil)
		switch {
		case status == C.CL_QUEUED:
			return CommandExecStatusQueued, toError(err)
		case status == C.CL_SUBMITTED:
			return CommandExecStatusSubmitted, toError(err)
		case status == C.CL_RUNNING:
			return CommandExecStatusRunning, toError(err)
		case status == C.CL_COMPLETE:
			return CommandExecStatusComplete, toError(err)
		default:
			return -1, toError(err)
		}
	}
	return -1, toError(C.CL_INVALID_EVENT)
}

func (e *Event) GetReferenceCount() (int, error) {
	if e.clEvent != nil {
		var outCount C.cl_uint
		err := C.clGetEventInfo(e.clEvent, C.CL_EVENT_REFERENCE_COUNT, C.size_t(unsafe.Sizeof(outCount)), unsafe.Pointer(&outCount), nil)
		return int(outCount), toError(err)
	}
	return 0, toError(C.CL_INVALID_EVENT)
}

func (ctx *Context) CreateUserEvent() (*Event, error) {
	var err C.cl_int
	clEvent := C.clCreateUserEvent(ctx.clContext, &err)
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	return newEvent(clEvent), nil
}

func (ev *Event) SetUserEventStatus(status CommandExecStatus) error {
	return toError(C.clSetUserEventStatus(ev.clEvent, (C.cl_int)(status)))
}

func (ev *Event) SetEventCallback(status CommandExecStatus, user_data unsafe.Pointer) error {
	return toError(C.CLSetEventCallback(ev.clEvent, (C.cl_int)(status), user_data))
}

// A synchronization point that enqueues a barrier operation.
func (q *CommandQueue) EnqueueBarrierWithWaitList(eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	err := toError(C.clEnqueueBarrierWithWaitList(q.clQueue, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

// Enqueues a marker command which waits for either a list of events to complete, or all previously enqueued commands to complete.
func (q *CommandQueue) EnqueueMarkerWithWaitList(eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	err := toError(C.clEnqueueMarkerWithWaitList(q.clQueue, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}
