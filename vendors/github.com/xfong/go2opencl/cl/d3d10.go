//go:build dx || gl_dx || dx_gl
// +build dx gl_dx dx_gl

package cl

/*
#include "./opencl.h"

static cl_int (*clGetDeviceIDsFromD3D10) (cl_platform_id, cl_d3d10_device_source_khr, void *, cl_d3d10_device_set_khr, cl_uint, cl_device_id *, cl_uint *);
static cl_mem (*clCreateFromD3D10Buffer) (cl_context, cl_mem_flags, ID3D10Buffer *, cl_int *);
static cl_mem (*clCreateFromD3D10Texture2D) (cl_context, cl_mem_flags, ID3D10Texture2D *, UINT, cl_int *);
static cl_mem (*clCreateFromD3D10Texture3D) (cl_context, cl_mem_flags, ID3D10Texture3D *, UINT, cl_int *);
static cl_int (*clEnqueueAcquireD3D10Objects) (cl_command_queue, cl_uint, const cl_mem *, cl_uint, const cl_event *, cl_event *);
static cl_int (*clEnqueueReleaseD3D10Objects) (cl_command_queue, cl_uint, const cl_mem *, cl_uint, const cl_event *, cl_event *);

static void SetupD3D10Sharing(cl_platform_id platform) {
	clGetDeviceIDsFromD3D10KHR_fn tmpPtr0 = NULL;
	tmpPtr0 = (clGetDeviceIDsFromD3D10KHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clGetDeviceIDsFromD3D10KHR");
	clGetDeviceIDsFromD3D10 = tmpPtr0;

	clCreateFromD3D10BufferKHR_fn tmpPtr1 = NULL;
	tmpPtr1 = (clCreateFromD3D10BufferKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromD3D10BufferKHR");
	clCreateFromD3D10Buffer = tmpPtr1;


	clCreateFromD3D10Texture2DKHR_fn tmpPtr2 = NULL;
	tmpPtr2 = (clCreateFromD3D10Texture2DKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromD3D10Texture2DKHR");
	clCreateFromD3D10Texture2D = tmpPtr2;

	clCreateFromD3D10Texture3DKHR_fn tmpPtr3 = NULL;
	tmpPtr3 = (clCreateFromD3D10Texture3DKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromD3D10Texture3DKHR");
	clCreateFromD3D10Texture3D = tmpPtr3;

	clEnqueueAcquireD3D10ObjectsKHR_fn tmpPtr4 = NULL;
	tmpPtr4 = (clEnqueueAcquireD3D10ObjectsKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clEnqueueAcquireD3D10ObjectsKHR");
	clEnqueueAcquireD3D10Objects = tmpPtr4;

	clEnqueueReleaseD3D10ObjectsKHR_fn tmpPtr5 = NULL;
	tmpPtr5 = (clEnqueueReleaseD3D10ObjectsKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clEnqueueReleaseD3D10ObjectsKHR");
	clEnqueueReleaseD3D10Objects = tmpPtr5;
}

static cl_int CLGetDeviceIDsFromD3D10(cl_platform_id platform, cl_d3d10_device_source_khr device_resource, void *d3d_object, cl_d3d10_device_set_khr d3d_device_set, cl_uint num_entries, cl_device_id *devices, cl_uint *num_devices) {
	return clGetDeviceIDsFromD3D10(platform, device_resource, d3d_object, d3d_device_set, num_entries, devices, num_devices);
}

static cl_mem CLCreateFromD3D10Buffer(cl_context context, cl_mem_flags flags, ID3D10Buffer *d3d10buffer, cl_int *errcode) {
	return clCreateFromD3D10Buffer(context, flags, d3d10buffer, errcode);
}

static cl_mem CLCreateFromD3D10Texture2D(cl_context context, cl_mem_flags flags, ID3D10Texture2D *d3d10texture2d, UINT subresource, cl_int *errcode) {
	return clCreateFromD3D10Texture2D(context, flags, d3d10texture2d, subresource, errcode);
}

static cl_mem CLCreateFromD3D10Texture3D(cl_context context, cl_mem_flags flags, ID3D10Texture3D *d3d10texture3d, UINT subresource, cl_int *errcode) {
	return clCreateFromD3D10Texture3D(context, flags, d3d10texture3d, subresource, errcode);
}

static cl_int CLEnqueueAcquireD3D10Objects(cl_command_queue com_queue, cl_uint num_obj, cl_mem *mem_obj_list, cl_uint num_events_in_wait_list, const cl_event *event_wait_list, cl_event *event) {
	return clEnqueueAcquireD3D10Objects(com_queue, num_obj, mem_obj_list, num_events_in_wait_list, event_wait_list, event);
}

static cl_int CLEnqueueReleaseD3D10Objects(cl_command_queue com_queue, cl_uint num_obj, cl_mem *mem_obj_list, cl_uint num_events_in_wait_list, const cl_event *event_wait_list, cl_event *event) {
	return clEnqueueReleaseD3D10Objects(com_queue, num_obj, mem_obj_list, num_events_in_wait_list, event_wait_list, event);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

//////////////// Basic Types ////////////////
var (
	ErrInvalidD3D10Device           = errors.New("cl: Invalid D3D10 Device")
	ErrInvalidD3D10Resource         = errors.New("cl: Invalid D3D10 Resource")
	ErrD3D10ResourceAlreadyAcquired = errors.New("cl: D3D10 Resource Already Acquired")
	ErrD3D10ResourceNotAcquired     = errors.New("cl: D3D10 Resource Not Acquired")
)

const (
	ContextD3D10PreferredSharedResources ContextInfo = C.CL_CONTEXT_D3D10_PREFER_SHARED_RESOURCES_KHR

	ContextD3D10Device ContextPropertiesId = C.CL_CONTEXT_D3D10_DEVICE_KHR

	CommandAcquireD3D10Objects CommandType = C.CL_COMMAND_ACQUIRE_D3D10_OBJECTS_KHR
	CommandReleaseD3D10Objects CommandType = C.CL_COMMAND_RELEASE_D3D10_OBJECTS_KHR
)

type CLD3D10DeviceSourceKHR int

const (
	CLD3D10Device      CLD3D10DeviceSourceKHR = C.CL_D3D10_DEVICE_KHR
	CLD3D10DXGIAdapter CLD3D10DeviceSourceKHR = C.CL_D3D10_DXGI_ADAPTER_KHR
)

type CLD3D10DeviceSetKHR int

const (
	CLD3D10PreferredDevices CLD3D10DeviceSetKHR = C.CL_PREFERRED_DEVICES_FOR_D3D10_KHR
	CLD3D10AllDevices       CLD3D10DeviceSetKHR = C.CL_ALL_DEVICES_FOR_D3D10_KHR
)

//////////////// Basic Functions ////////////////
func init() {
	errorMap[C.CL_INVALID_D3D10_DEVICE_KHR] = ErrInvalidD3D10Device
	errorMap[C.CL_INVALID_D3D10_RESOURCE_KHR] = ErrInvalidD3D10Resource
	errorMap[C.CL_D3D10_RESOURCE_ALREADY_ACQUIRED_KHR] = ErrD3D10ResourceAlreadyAcquired
	errorMap[C.CL_D3D10_RESOURCE_NOT_ACQUIRED_KHR] = ErrD3D10ResourceNotAcquired

	d3d10_sharing_ext = true
	getD3D10CommandType = D3D10StatusToCommandType
}

func D3D10StatusToCommandType(status C.cl_command_type) (bool, CommandType) {
	switch status {
	case C.CL_COMMAND_ACQUIRE_D3D10_OBJECTS_KHR:
		return true, CommandAcquireD3D10Objects
	case C.CL_COMMAND_RELEASE_D3D10_OBJECTS_KHR:
		return true, CommandReleaseD3D10Objects
	default:
		return false, -1
	}
}

//////////////// Abstract Functions ////////////////
func (p *Platform) SetupD3D10Sharing() {
	C.SetupD3D10Sharing(p.id)
}

func (p *Platform) GetDeviceIDsFromD3D10(D3D10DeviceSrc CLD3D10DeviceSourceKHR, D3D10Obj unsafe.Pointer, D3D10DeviceSet CLD3D10DeviceSetKHR, num_devices int) ([]*Device, error) {
	var device_id_tmp []C.cl_device_id
	defer C.free(device_id_tmp)
	var device_count C.cl_uint
	defer C.free(device_count)
	err := C.CLGetDeviceIDsFromD3D10(p.id, (C.cl_d3d10_device_source_khr)(D3D10DeviceSrc), D3D10Obj, (C.cl_d3d10_device_set_khr)(D3D10DeviceSet), (C.cl_uint)(num_devices), &device_id_tmp[0], &device_count)
	go_count := int(device_count)
	outDeviceList := make([]*Device, go_count)
	for i := 0; i < go_count; i++ {
		outDeviceList[i].id = device_id_tmp[i]
	}
	return outDeviceList, toError(err)
}

func (ctx *Context) CreateFromD3D10Buffer(flag MemFlag, src unsafe.Pointer) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromD3D10Buffer(ctx.clContext, (C.cl_mem_flags)(flag), (*C.ID3D10Buffer)(src), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromD3D10BufferKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (ctx *Context) CreateFromD3D10Texture2D(flag MemFlag, src unsafe.Pointer, subResource int) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromD3D10Texture2D(ctx.clContext, (C.cl_mem_flags)(flag), (*C.ID3D10Texture2D)(src), (C.UINT)(subResource), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromD3D10BufferKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (ctx *Context) CreateFromD3D10Texture3D(flag MemFlag, src unsafe.Pointer, subResource int) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromD3D10Texture3D(ctx.clContext, (C.cl_mem_flags)(flag), (*C.ID3D10Texture3D)(src), (C.UINT)(subResource), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromD3D10BufferKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (q *CommandQueue) EnqueueAcquireD3D10Objects(memObj []*MemObject, eventWaitList []*Event) (*Event, error) {
	memList := make([]C.cl_mem, len(memObj))
	for i, ptr := range memObj {
		tmpObj := *ptr
		memList[i] = tmpObj.clMem
	}
	var event C.cl_event
	err := C.CLEnqueueAcquireD3D10Objects(q.clQueue, (C.cl_uint)(len(memObj)), &memList[0], (C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}

func (q *CommandQueue) EnqueueReleaseD3D10Objects(memObj []*MemObject, eventWaitList []*Event) (*Event, error) {
	memList := make([]C.cl_mem, len(memObj))
	for i, ptr := range memObj {
		tmpObj := *ptr
		memList[i] = tmpObj.clMem
	}
	var event C.cl_event
	err := C.CLEnqueueReleaseD3D10Objects(q.clQueue, (C.cl_uint)(len(memObj)), &memList[0], (C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}

func (b *MemObject) GetD3D10Resource() (*C.ID3D10Resource, error) {
	if b.clMem != nil {
		var val C.ID3D10Resource
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_D3D10_RESOURCE_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return nil, toError(err)
		}
		return &val, nil
	}
	return nil, toError(C.CL_INVALID_MEM_OBJECT)
}

func (image_desc *ImageDescription) GetD3D10Subresource() (*C.ID3D10Resource, error) {
	if image_desc.Buffer != nil {
		var val C.ID3D10Resource
		err := C.clGetImageInfo(image_desc.Buffer.clMem, C.CL_IMAGE_D3D10_SUBRESOURCE_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return nil, toError(err)
		}
		return &val, nil
	}
	return nil, toError(C.CL_INVALID_MEM_OBJECT)
}
