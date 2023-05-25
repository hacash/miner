//go:build dx || gl_dx || dx_gl
// +build dx gl_dx dx_gl

package cl

/*
#include "./opencl.h"

static cl_int (*clGetDeviceIDsFromDX9MediaAdapter) (cl_platform_id, cl_uint, cl_dx9_media_adapter_type_khr *, void *, cl_dx9_media_adapter_set_khr, cl_uint, cl_device_id *, cl_uint *);
static cl_mem (*clCreateFromDX9MediaSurface) (cl_context, cl_mem_flags, cl_dx9_media_adapter_type_khr, void *, cl_uint, cl_int *);
static cl_int (*clEnqueueAcquireDX9MediaSurfaces) (cl_command_queue, cl_uint, const cl_mem *, cl_uint, const cl_event *, cl_event *);
static cl_int (*clEnqueueReleaseDX9MediaSurfaces) (cl_command_queue, cl_uint, const cl_mem *, cl_uint, const cl_event *, cl_event *);

static void SetupDX9MediaSharing(cl_platform_id platform) {
	clGetDeviceIDsFromDX9MediaAdapterKHR_fn tmpPtr0 = NULL;
	tmpPtr0 = (clGetDeviceIDsFromDX9MediaAdapterKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clGetDeviceIDsFromDX9MediaAdapterKHR");
	clGetDeviceIDsFromDX9MediaAdapter = tmpPtr0;

	clCreateFromDX9MediaSurfaceKHR_fn tmpPtr1 = NULL;
	tmpPtr1 = (clCreateFromDX9MediaSurfaceKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateFromDX9MediaSurfaceKHR");
	clCreateFromDX9MediaSurface = tmpPtr1;

	clEnqueueAcquireDX9MediaSurfacesKHR_fn tmpPtr2 = NULL;
	tmpPtr2 = (clEnqueueAcquireDX9MediaSurfacesKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clEnqueueAcquireDX9MediaSurfacesKHR");
	clEnqueueAcquireDX9MediaSurfaces = tmpPtr2;

	clEnqueueReleaseDX9MediaSurfacesKHR_fn tmpPtr3 = NULL;
	tmpPtr3 = (clEnqueueReleaseDX9MediaSurfacesKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clEnqueueReleaseDX9MediaSurfacesKHR");
	clEnqueueReleaseDX9MediaSurfaces = tmpPtr3;
}

static cl_int CLGetDeviceIDsFromDX9MediaAdapter(cl_platform_id platform, cl_uint num_adapters, cl_dx9_media_adapter_type_khr *media_adapters_type, void *media_adapters, cl_dx9_media_adapter_set_khr media_adapter_set, cl_uint num_entries, cl_device_id *devices, cl_uint *num_devices) {
	return clGetDeviceIDsFromDX9MediaAdapter(platform, num_adapters, media_adapters_type, media_adapters, media_adapter_set, num_entries, devices, num_devices);
}

static cl_mem CLCreateFromDX9MediaSurface(cl_context context, cl_mem_flags flags, cl_dx9_media_adapter_type_khr media_adapter_type, void * surface_info, cl_uint plane, cl_int *errcode) {
	return clCreateFromDX9MediaSurface(context, flags, media_adapter_type, surface_info, plane, errcode);
}

static cl_int CLEnqueueAcquireDX9MediaSurfaces(cl_command_queue com_queue, cl_uint num_obj, cl_mem *mem_obj_list, cl_uint num_events_in_wait_list, const cl_event *event_wait_list, cl_event *event) {
	return clEnqueueAcquireDX9MediaSurfaces(com_queue, num_obj, mem_obj_list, num_events_in_wait_list, event_wait_list, event);
}

static cl_int CLEnqueueReleaseDX9MediaSurfaces(cl_command_queue com_queue, cl_uint num_obj, cl_mem *mem_obj_list, cl_uint num_events_in_wait_list, const cl_event *event_wait_list, cl_event *event) {
	return clEnqueueReleaseDX9MediaSurfaces(com_queue, num_obj, mem_obj_list, num_events_in_wait_list, event_wait_list, event);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// ////////////// Basic Types ////////////////
const (
	CommandAcquireDX9Objects CommandType = C.CL_COMMAND_ACQUIRE_DX9_MEDIA_SURFACES_KHR
	CommandReleaseDX9Objects CommandType = C.CL_COMMAND_RELEASE_DX9_MEDIA_SURFACES_KHR

	ContextD3D9Adapter   ContextPropertiesId = C.CL_CONTEXT_ADAPTER_D3D9_KHR
	ContextD3D9EXAdapter ContextPropertiesId = C.CL_CONTEXT_ADAPTER_D3D9EX_KHR
	ContextDXVAAdapter   ContextPropertiesId = C.CL_CONTEXT_ADAPTER_DXVA_KHR
)

type CLDX9AdapterType int

var (
	ErrInvalidDX9MediaAdapter    = errors.New("cl: Invalid DX9 Media Adapter")
	ErrInvalidDX9MediaSurface    = errors.New("cl: Invalid DX9 Media Surface")
	ErrDX9SurfaceAlreadyAcquired = errors.New("cl: DX9 Surface Already Acquired")
	ErrDX9SurfaceNotAcquired     = errors.New("cl: DX9 Surface Not Acquired")
)

const (
	CLDX9Adapter   CLDX9AdapterType = C.CL_ADAPTER_D3D9_KHR
	CLDX9EXAdapter CLDX9AdapterType = C.CL_ADAPTER_D3D9EX_KHR
	CLDXVAAdapter  CLDX9AdapterType = C.CL_ADAPTER_DXVA_KHR
)

type CLDX9DeviceSetKHR int

const (
	CLDX9PreferredDevices CLDX9DeviceSetKHR = C.CL_PREFERRED_DEVICES_FOR_DX9_MEDIA_ADAPTER_KHR
	CLDX9AllDevices       CLDX9DeviceSetKHR = C.CL_ALL_DEVICES_FOR_DX9_MEDIA_ADAPTER_KHR
)

// ////////////// Basic Functions ////////////////
func init() {
	errorMap[C.CL_INVALID_DX9_MEDIA_ADAPTER_KHR] = ErrInvalidDX9MediaAdapter
	errorMap[C.CL_INVALID_DX9_MEDIA_SURFACE_KHR] = ErrInvalidDX9MediaSurface
	errorMap[C.CL_DX9_MEDIA_SURFACE_ALREADY_ACQUIRED_KHR] = ErrDX9SurfaceAlreadyAcquired
	errorMap[C.CL_DX9_MEDIA_SURFACE_NOT_ACQUIRED_KHR] = ErrDX9SurfaceNotAcquired

	dx9_sharing_ext = true
	getDX9CommandType = DX9StatusToCommandType
}

func DX9StatusToCommandType(status C.cl_command_type) (bool, CommandType) {
	switch status {
	case C.CL_COMMAND_ACQUIRE_DX9_MEDIA_SURFACES_KHR:
		return true, CommandAcquireDX9Objects
	case C.CL_COMMAND_RELEASE_DX9_MEDIA_SURFACES_KHR:
		return true, CommandReleaseDX9Objects
	default:
		return false, -1
	}
}

// ////////////// Abstract Functions ////////////////
func (p *Platform) SetupDX9Sharing() {
	C.SetupDX9MediaSharing(p.id)
}

func (p *Platform) GetDeviceIDsFromDX9MediaAdapter(num_adapters int, media_adapter_types []CLDX9AdapterType,
	media_adapters unsafe.Pointer, media_adapter_set CLDX9DeviceSetKHR) ([]*Device, error) {
	var device_id_tmp []C.cl_device_id
	defer C.free(device_id_tmp)
	var device_count C.cl_uint
	defer C.free(device_count)
	tmpAdapterTypeList := make([]C.cl_dx9_media_adapter_type_khr, len(media_adapter_types))
	defer C.free(tmpAdapterTypeList)
	for i, type_val := range media_adapter_types {
		tmpAdapterTypeList[i] = (C.cl_dx9_media_adapter_type_khr)(type_val)
	}
	err := C.CLGetDeviceIDsFromDX9MediaAdapter(p.id, (C.cl_uint)(num_adapters), &tmpAdapterTypeList[0], media_adapters,
		(C.cl_dx9_media_adapter_set_khr)(media_adapter_set), 1, &device_id_tmp[0], &device_count)
	if toError(err) != nil {
		return nil, toError(err)
	}
	err = C.CLGetDeviceIDsFromDX9MediaAdapter(p.id, (C.cl_uint)(num_adapters), &tmpAdapterTypeList[0], media_adapters,
		(C.cl_dx9_media_adapter_set_khr)(media_adapter_set), device_count, &device_id_tmp[0], nil)
	if toError(err) != nil {
		return nil, toError(err)
	}
	go_count := int(device_count)
	outDeviceList := make([]*Device, go_count)
	for i := 0; i < go_count; i++ {
		outDeviceList[i].id = device_id_tmp[i]
	}
	return outDeviceList, toError(err)
}

func (ctx *Context) CreateFromDX9MediaSurface(flag MemFlag, dx9_adapter_type CLDX9AdapterType,
	surface_info unsafe.Pointer, plane int) (*MemObject, error) {
	var err C.cl_int
	defer C.free(err)
	memObj := C.CLCreateFromDX9MediaSurface(ctx.clContext, (C.cl_mem_flags)(flag), (C.cl_dx9_media_adapter_type_khr)(dx9_adapter_type), surface_info, (C.cl_uint)(plane), &err)
	tmpBuf := &MemObject{clMem: memObj, size: 0}
	bufSize, sizeErr := tmpBuf.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromDX9MediaSurfaceKHR \n")
		return nil, sizeErr
	}
	return &MemObject{clMem: memObj, size: bufSize}, toError(err)
}

func (q *CommandQueue) EnqueueAcquireDX9MediaSurfaces(memObj []*MemObject, eventWaitList []*Event) (*Event, error) {
	memList := make([]C.cl_mem, len(memObj))
	for i, ptr := range memObj {
		tmpObj := *ptr
		memList[i] = tmpObj.clMem
	}
	var event C.cl_event
	err := C.CLEnqueueAcquireDX9MediaSurfaces(q.clQueue, (C.cl_uint)(len(memObj)), &memList[0],
		(C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList),
		&event)
	return newEvent(event), toError(err)
}

func (q *CommandQueue) EnqueueReleaseDX9MediaSurfaces(memObj []*MemObject, eventWaitList []*Event) (*Event, error) {
	memList := make([]C.cl_mem, len(memObj))
	for i, ptr := range memObj {
		tmpObj := *ptr
		memList[i] = tmpObj.clMem
	}
	var event C.cl_event
	err := C.CLEnqueueReleaseDX9MediaSurfaces(q.clQueue,
		(C.cl_uint)(len(memObj)), &memList[0], (C.cl_uint)(len(eventWaitList)),
		eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}

func (b *MemObject) GetDX9MediaAdapterType() (CLDX9AdapterType, error) {
	if b.clMem != nil {
		var val C.cl_dx9_media_adapter_type_khr
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_DX9_MEDIA_ADAPTER_TYPE_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return -1, toError(err)
		}
		switch val {
		default:
			return -1, toError(err)
		case C.CL_CONTEXT_ADAPTER_D3D9_KHR:
			return CLDX9Adapter, nil
		case C.CL_CONTEXT_ADAPTER_D3D9EX_KHR:
			return CLDX9EXAdapter, nil
		case C.CL_CONTEXT_ADAPTER_DXVA_KHR:
			return CLDXVAAdapter, nil
		}
	}
	return -1, toError(C.CL_INVALID_MEM_OBJECT)
}

func (b *MemObject) GetDX9SurfaceInfo() (unsafe.Pointer, error) {
	if b.clMem != nil {
		var val C.cl_dx9_surface_info_khr
		err := C.clGetMemObjectInfo(b.clMem, C.CL_MEM_DX9_MEDIA_SURFACE_INFO_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return nil, toError(err)
		}
		return (unsafe.Pointer)(&val), nil
	}
	return nil, toError(C.CL_INVALID_MEM_OBJECT)
}

func (image_desc *ImageDescription) GetDX9MediaPlane() (int, error) {
	if image_desc.Buffer != nil {
		var val C.cl_uint
		err := C.clGetImageInfo(image_desc.Buffer.clMem, C.CL_IMAGE_DX9_MEDIA_PLANE_KHR, (C.size_t)(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
		if toError(err) != nil {
			return -1, toError(err)
		}
		return int(val), nil
	}
	return -1, toError(C.CL_INVALID_MEM_OBJECT)
}
