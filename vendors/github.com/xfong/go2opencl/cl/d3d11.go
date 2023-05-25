//go:build dx || gl_dx || dx_gl
// +build dx gl_dx dx_gl

package cl

/*
#include "./opencl.h"

static cl_int (*clGetDeviceIDsFromD3D11) (cl_platform_id, cl_d3d11_device_source_khr, void *, cl_d3d11_device_set_khr, cl_uint, cl_device_id *, cl_uint *);
static cl_mem (*clCreateFromD3D11Buffer) (cl_context, cl_mem_flags, ID3D11Buffer *, cl_int *);
static cl_mem (*clCreateFromD3D11Texture2D) (cl_context, cl_mem_flags, ID3D11Texture2D *, UINT, cl_int *);
static cl_mem (*clCreateFromD3D11Texture3D) (cl_context, cl_mem_flags, ID3D11Texture3D *, UINT, cl_int *);
static cl_int (*clEnqueueAcquireD3D11Objects) (cl_command_queue, cl_uint, const cl_mem *, cl_uint, const cl_event *, cl_event *);
static cl_int (*clEnqueueReleaseD3D11Objects) (cl_command_queue, cl_uint, const cl_mem *, cl_uint, const cl_event *, cl_event *);

static void SetupD3D11Sharing(cl_platform_id platform) {
	clGetDeviceIDsFromD3D11KHR_fn tmpPtr0 = NULL;
	tmpPtr0 = (clGetDeviceIDsFromD3D11KHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clGetDeviceIDsFromD3D11KHR");
	clGetDeviceIDsFromD3D11 = tmpPtr0;

	clCreateFromD3D11BufferKHR_fn tmpPtr1 = NULL;
	tmpPtr1 = (clCreateFromD3D11BufferKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromD3D11BufferKHR");
	clCreateFromD3D11Buffer = tmpPtr1;


	clCreateFromD3D11Texture2DKHR_fn tmpPtr2 = NULL;
	tmpPtr2 = (clCreateFromD3D11Texture2DKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromD3D11Texture2DKHR");
	clCreateFromD3D11Texture2D = tmpPtr2;

	clCreateFromD3D11Texture3DKHR_fn tmpPtr3 = NULL;
	tmpPtr3 = (clCreateFromD3D11Texture3DKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromD3D11Texture3DKHR");
	clCreateFromD3D11Texture3D = tmpPtr3;

	clEnqueueAcquireD3D11ObjectsKHR_fn tmpPtr4 = NULL;
	tmpPtr4 = (clEnqueueAcquireD3D11ObjectsKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clEnqueueAcquireD3D11ObjectsKHR");
	clEnqueueAcquireD3D11Objects = tmpPtr4;

	clEnqueueReleaseD3D11ObjectsKHR_fn tmpPtr5 = NULL;
	tmpPtr5 = (clEnqueueReleaseD3D11ObjectsKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clEnqueueReleaseD3D11ObjectsKHR");
	clEnqueueReleaseD3D11Objects = tmpPtr5;
}

static cl_int CLGetDeviceIDsFromD3D11(cl_platform_id platform, cl_d3d11_device_source_khr device_resource, void *d3d_object, cl_d3d11_device_set_khr d3d_device_set, cl_uint num_entries, cl_device_id *devices, cl_uint *num_devices) {
	return clGetDeviceIDsFromD3D11(platform, device_resource, d3d_object, d3d_device_set, num_entries, devices, num_devices);
}

static cl_mem CLCreateFromD3D11Buffer(cl_context context, cl_mem_flags flags, ID3D11Buffer *d3d11buffer, cl_int *errcode) {
	return clCreateFromD3D11Buffer(context, flags, d3d11buffer, errcode);
}

static cl_mem CLCreateFromD3D11Texture2D(cl_context context, cl_mem_flags flags, ID3D11Texture2D *d3d11texture2d, UINT subresource, cl_int *errcode) {
	return clCreateFromD3D11Texture2D(context, flags, d3d11texture2d, subresource, errcode);
}

static cl_mem CLCreateFromD3D11Texture3D(cl_context context, cl_mem_flags flags, ID3D11Texture3D *d3d11texture3d, UINT subresource, cl_int *errcode) {
	return clCreateFromD3D11Texture3D(context, flags, d3d11texture3d, subresource, errcode);
}

static cl_int CLEnqueueAcquireD3D11Objects(cl_command_queue com_queue, cl_uint num_obj, cl_mem *mem_obj_list, cl_uint num_events_in_wait_list, const cl_event *event_wait_list, cl_event *event) {
	return clEnqueueAcquireD3D11Objects(com_queue, num_obj, mem_obj_list, num_events_in_wait_list, event_wait_list, event);
}

static cl_int CLEnqueueReleaseD3D11Objects(cl_command_queue com_queue, cl_uint num_obj, cl_mem *mem_obj_list, cl_uint num_events_in_wait_list, const cl_event *event_wait_list, cl_event *event) {
	return clEnqueueReleaseD3D11Objects(com_queue, num_obj, mem_obj_list, num_events_in_wait_list, event_wait_list, event);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// ////////////// Basic Types ////////////////
var (
	ErrInvalidD3D11Device           = errors.New("cl: Invalid D3D11 Device")
	ErrInvalidD3D11Resource         = errors.New("cl: Invalid D3D11 Resource")
	ErrD3D11ResourceAlreadyAcquired = errors.New("cl: D3D11 Resource Already Acquired")
	ErrD3D11ResourceNotAcquired     = errors.New("cl: D3D11 Resource Not Acquired")
)

const (
	ContextD3D11PreferredSharedResources ContextInfo = C.CL_CONTEXT_D3D11_PREFER_SHARED_RESOURCES_KHR

	ContextD3D11Device ContextPropertiesId = C.CL_CONTEXT_D3D11_DEVICE_KHR

	CommandAcquireD3D11Objects CommandType = C.CL_COMMAND_ACQUIRE_D3D11_OBJECTS_KHR
	CommandReleaseD3D11Objects CommandType = C.CL_COMMAND_RELEASE_D3D11_OBJECTS_KHR
)

type CLD3D11DeviceSourceKHR int

const (
	CLD3D11Device      CLD3D11DeviceSourceKHR = C.CL_D3D11_DEVICE_KHR
	CLD3D11DXGIAdapter CLD3D11DeviceSourceKHR = C.CL_D3D11_DXGI_ADAPTER_KHR
)

type CLD3D11DeviceSetKHR int

const (
	CLD3D11PreferredDevices CLD3D11DeviceSetKHR = C.CL_PREFERRED_DEVICES_FOR_D3D11_KHR
	CLD3D11AllDevices       CLD3D11DeviceSetKHR = C.CL_ALL_DEVICES_FOR_D3D11_KHR
)

// ////////////// Basic Functions ////////////////
func init() {
	errorMap[C.CL_INVALID_D3D11_DEVICE_KHR] = ErrInvalidD3D11Device
	errorMap[C.CL_INVALID_D3D11_RESOURCE_KHR] = ErrInvalidD3D11Resource
	errorMap[C.CL_D3D11_RESOURCE_ALREADY_ACQUIRED_KHR] = ErrD3D11ResourceAlreadyAcquired
	errorMap[C.CL_D3D11_RESOURCE_NOT_ACQUIRED_KHR] = ErrD3D11ResourceNotAcquired

	d3d11_sharing_ext = true
	getD3D11CommandType = D3D11StatusToCommandType
}

func D3D11StatusToCommandType(status C.cl_command_type) (bool, CommandType) {
	switch status {
	case C.CL_COMMAND_ACQUIRE_D3D11_OBJECTS_KHR:
		return true, CommandAcquireD3D11Objects
	case C.CL_COMMAND_RELEASE_D3D11_OBJECTS_KHR:
		return true, CommandReleaseD3D11Objects
	default:
		return false, -1
	}
}

// ////////////// Abstract Functions ////////////////
func (p *Platform) SetupD3D11Sharing() {
	C.SetupD3D11Sharing(p.id)
}

func (p *Platform) GetDeviceIDsFromD3D11(D3D11DeviceSrc CLD3D11DeviceSourceKHR, D3D11Obj unsafe.Pointer, D3D11DeviceSet CLD3D11DeviceSetKHR, num_devices int) ([]*Device, error) {
	var device_id_tmp []C.cl_device_id
	defer C.free(device_id_tmp)
	var device_count C.cl_uint
	defer C.free(device_count)
	err := C.CLGetDeviceIDsFromD3D11(p.id, (C.cl_d3d11_device_source_khr)(D3D11DeviceSrc), D3D11Obj, (C.cl_d3d11_device_set_khr)(D3D11DeviceSet), (C.cl_uint)(num_devices), &device_id_tmp[0], &device_count)
	go_count := int(device_count)
	outDeviceList := make([]*Device, go_count)
	for i := 0; i < go_count; i++ {
		outDeviceList[i].id = device_id_tmp[i]
	}
	return outDeviceList, toError(err)
}

func (ctx *Context) CreateFromD3D11Buffer(flag MemFlag, src unsafe.Pointer) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromD3D11Buffer(ctx.clContext, (C.cl_mem_flags)(flag), (*C.ID3D11Buffer)(src), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromD3D11BufferKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (ctx *Context) CreateFromD3D11Texture2D(flag MemFlag, src unsafe.Pointer, subResource int) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromD3D11Texture2D(ctx.clContext, (C.cl_mem_flags)(flag), (*C.ID3D11Texture2D)(src), (C.UINT)(subResource), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromD3D11BufferKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (ctx *Context) CreateFromD3D11Texture3D(flag MemFlag, src unsafe.Pointer, subResource int) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromD3D11Texture3D(ctx.clContext, (C.cl_mem_flags)(flag), (*C.ID3D11Texture3D)(src), (C.UINT)(subResource), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromD3D11BufferKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (q *CommandQueue) EnqueueAcquireD3D11Objects(memObj []*MemObject, eventWaitList []*Event) (*Event, error) {
	memList := make([]C.cl_mem, len(memObj))
	for i, ptr := range memObj {
		tmpObj := *ptr
		memList[i] = tmpObj.clMem
	}
	var event C.cl_event
	err := C.CLEnqueueAcquireD3D11Objects(q.clQueue, (C.cl_uint)(len(memObj)), &memList[0], (C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}

func (q *CommandQueue) EnqueueReleaseD3D11Objects(memObj []*MemObject, eventWaitList []*Event) (*Event, error) {
	memList := make([]C.cl_mem, len(memObj))
	for i, ptr := range memObj {
		tmpObj := *ptr
		memList[i] = tmpObj.clMem
	}
	var event C.cl_event
	err := C.CLEnqueueReleaseD3D11Objects(q.clQueue, (C.cl_uint)(len(memObj)), &memList[0], (C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}

func (b *MemObject) GetD3D11Resource() (*C.ID3D11Resource, error) {
	if b.clMem != nil {
		var val C.ID3D11Resource
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_D3D11_RESOURCE_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return nil, toError(err)
		}
		return &val, nil
	}
	return nil, toError(C.CL_INVALID_MEM_OBJECT)
}

func (image_desc *ImageDescription) GetD3D11Subresource() (*C.ID3D11Resource, error) {
	if image_desc.Buffer != nil {
		var val C.ID3D11Resource
		err := C.clGetImageInfo(image_desc.Buffer.clMem, C.CL_IMAGE_D3D11_SUBRESOURCE_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return nil, toError(err)
		}
		return &val, nil
	}
	return nil, toError(C.CL_INVALID_MEM_OBJECT)
}
