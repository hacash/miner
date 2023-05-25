package cl

/*
#include "./opencl.h"

extern void go_ctx_notify(char *errinfo, void *private_info, int cb, void *user_data);
static void CL_CALLBACK c_ctx_notify(const char *errinfo, const void *private_info, size_t cb, void *user_data) {
        go_ctx_notify((char *)errinfo, (void *)private_info, cb, user_data);
}

static cl_context CLCreateContext(      const cl_context_properties *   properties,
                                                        cl_uint                                 num_devices,
                                                        const cl_device_id *                    devices,
                                                        void *                                  user_data,
                                                        cl_int *                                errcode_ret){
        return clCreateContext(properties, num_devices, devices, c_ctx_notify, user_data, errcode_ret);
}

static cl_context CLCreateContextFromType(      const cl_context_properties *   properties,
                                                                        cl_device_type                                  device_type,
                                                                        void *                                  user_data,
                                                                        cl_int *                                errcode_ret){
    return clCreateContextFromType(properties, device_type, c_ctx_notify, user_data, errcode_ret);
}

static cl_context_properties platform_id_convert(cl_platform_id id) {
        return (cl_context_properties)(id);
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

////////////////// Basic Types ////////////////
type ContextInfo int

const (
	ContextReferenceCount ContextInfo = C.CL_CONTEXT_REFERENCE_COUNT
	ContextDevices        ContextInfo = C.CL_CONTEXT_DEVICES
	ContextNumDevices     ContextInfo = C.CL_CONTEXT_NUM_DEVICES
	ContextProperties     ContextInfo = C.CL_CONTEXT_PROPERTIES
)

type ContextPropertiesId int

const (
	ContextPlatform        ContextPropertiesId = C.CL_CONTEXT_PLATFORM
	ContextInteropUserSync ContextPropertiesId = C.CL_CONTEXT_INTEROP_USER_SYNC
)

////////////////// Abstract Types ////////////////
type Context struct {
	clContext C.cl_context
	devices   []*Device
}

////////////////// Golang Types ////////////////
type CLContext C.cl_context
type CLContextProperties C.cl_context_properties

////////////////// Supporting Types ////////////////
type CL_ctx_notify func(errinfo string, private_info unsafe.Pointer, cb int, user_data unsafe.Pointer)

var ctx_notify map[unsafe.Pointer]CL_ctx_notify

////////////////// Basic Functions ////////////////
func init() {
	ctx_notify = make(map[unsafe.Pointer]CL_ctx_notify)
}

//export go_ctx_notify
func go_ctx_notify(errinfo *C.char, private_info unsafe.Pointer, cb C.int, user_data unsafe.Pointer) {
	var c_user_data []unsafe.Pointer
	c_user_data = *(*[]unsafe.Pointer)(user_data)
	ctx_notify[c_user_data[1]](C.GoString(errinfo), private_info, int(cb), c_user_data[0])
}

func releaseContext(c *Context) {
	if c.clContext != nil {
		C.clReleaseContext(c.clContext)
		c.clContext = nil
	}
}

func retainContext(c *Context) {
	if c.clContext != nil {
		C.clRetainContext(c.clContext)
	}
}

func CreateContext(devices []*Device) (*Context, error) {
	clContext, err := CreateContextUnsafe(nil, devices, nil, nil)
	return clContext, err
}

func CreateContextUnsafe(properties *C.cl_context_properties, devices []*Device, pfn_notify CL_ctx_notify, user_data unsafe.Pointer) (*Context, error) {
	deviceIds := buildDeviceIdList(devices)
	var err C.cl_int
	var clContext C.cl_context
	if pfn_notify != nil {
		var c_user_data []unsafe.Pointer
		c_user_data = make([]unsafe.Pointer, 2)
		c_user_data[0] = user_data
		c_user_data[1] = unsafe.Pointer(&pfn_notify)

		ctx_notify[c_user_data[1]] = pfn_notify

		clContext = C.CLCreateContext(properties, C.cl_uint(len(devices)), &deviceIds[0], unsafe.Pointer(&c_user_data), &err)
	} else {
		clContext = C.clCreateContext(properties, C.cl_uint(len(devices)), &deviceIds[0], nil, nil, &err)
	}
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	if clContext == nil {
		return nil, ErrUnknown
	}
	context := &Context{clContext: clContext, devices: devices}
	runtime.SetFinalizer(context, releaseContext)
	return context, nil
}

func CreateContextFromTypeUnsafe(properties *C.cl_context_properties, device_type C.cl_device_type, pfn_notify CL_ctx_notify, user_data unsafe.Pointer) (*Context, error) {
	var err C.cl_int
	var clContext C.cl_context
	if pfn_notify != nil {
		var c_user_data []unsafe.Pointer
		c_user_data = make([]unsafe.Pointer, 2)
		c_user_data[0] = user_data
		c_user_data[1] = unsafe.Pointer(&pfn_notify)

		ctx_notify[c_user_data[1]] = pfn_notify

		clContext = C.CLCreateContextFromType(properties, device_type, unsafe.Pointer(&c_user_data), &err)
	} else {
		clContext = C.clCreateContextFromType(properties, device_type, nil, nil, &err)
	}
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	if clContext == nil {
		return nil, ErrUnknown
	}
	contextTmp := &Context{clContext: clContext, devices: nil}
	cDevices, errD := contextTmp.GetDevices()
	if errD != nil {
		runtime.SetFinalizer(contextTmp, releaseContext)
		return contextTmp, toError(err)
	}
	context := &Context{clContext: clContext, devices: cDevices}
	runtime.SetFinalizer(context, releaseContext)
	return context, nil
}

////////////////// Abstract Functions ////////////////
func (ctx *Context) Release() {
	releaseContext(ctx)
}

func (ctx *Context) Retain() {
	retainContext(ctx)
}

func (ctx *Context) GetReferenceCount() (int, error) {
	if ctx.clContext != nil {
		var outCount C.cl_uint
		err := C.clGetContextInfo(ctx.clContext, C.cl_context_info(ContextReferenceCount), C.size_t(unsafe.Sizeof(outCount)), unsafe.Pointer(&outCount), nil)
		return int(outCount), toError(err)
	}
	return 0, toError(C.CL_INVALID_CONTEXT)
}

func (ctx *Context) GetDevices() ([]*Device, error) {
	if ctx.clContext != nil {
		var tmpCount C.cl_device_id
		var outDevices []C.cl_device_id
		var devCount C.size_t
		err := C.clGetContextInfo(ctx.clContext, C.cl_context_info(ContextDevices), C.size_t(unsafe.Sizeof(tmpCount)), unsafe.Pointer(&outDevices), &devCount)
		if int(devCount) != 0 {
			devPtr := make([]*Device, int(devCount))
			for i := range devPtr {
				devPtr[i].id = outDevices[i]
			}
			return devPtr, toError(err)
		}
		return nil, toError(err)
	}
	return nil, toError(C.CL_INVALID_CONTEXT)
}

func (ctx *Context) GetNumberOfDevices() (int, error) {
	if ctx.clContext != nil {
		var outCount C.cl_uint
		err := C.clGetContextInfo(ctx.clContext, C.cl_context_info(ContextNumDevices), C.size_t(unsafe.Sizeof(outCount)), unsafe.Pointer(&outCount), nil)
		return int(outCount), toError(err)
	}
	return 0, toError(C.CL_INVALID_CONTEXT)
}

func (ctx *Context) GetProperties() ([]CLContextProperties, error) {
	if ctx.clContext != nil {
		var tmpProperty CLContextProperties
		var tmpList []C.cl_context_properties
		var tmpCount C.size_t
		err := C.clGetContextInfo(ctx.clContext, C.cl_context_info(ContextProperties), C.size_t(unsafe.Sizeof(tmpProperty)), unsafe.Pointer(&tmpList), &tmpCount)
		if toError(err) == nil {
			if tmpCount == 0 {
				return nil, nil
			} else {
				var outList []CLContextProperties
				for i := 0; i < int(tmpCount/C.size_t(unsafe.Sizeof(tmpProperty))); i++ {
					outList[i] = (CLContextProperties)(tmpList[i])
				}
				return outList, nil
			}
		}
		return []CLContextProperties{}, toError(err)
	}
	return []CLContextProperties{}, toError(C.CL_INVALID_CONTEXT)
}

func (p *Platform) CreateContext(devList []*Device) (*Context, error) {
	if devList != nil {
		var properties []C.cl_context_properties
		defer C.free(unsafe.Pointer(&properties))
		properties = append(properties, C.CL_CONTEXT_PLATFORM)
		properties = append(properties, (C.platform_id_convert(p.id)))
		properties = append(properties, (C.cl_context_properties)(0))
		ctx, err := CreateContextUnsafe(&properties[0], devList, nil, nil)
		return ctx, err
	}
	return nil, toError(C.CL_INVALID_DEVICE)
}

func (p *Platform) CreateContextFromType(device_type DeviceType) (*Context, error) {
	if device_type == DeviceTypeCPU || device_type == DeviceTypeGPU || device_type == DeviceTypeAccelerator || device_type == DeviceTypeDefault || device_type == DeviceTypeAll {
		var properties []C.cl_context_properties
		defer C.free(unsafe.Pointer(&properties))
		properties = append(properties, C.CL_CONTEXT_PLATFORM)
		properties = append(properties, (C.platform_id_convert(p.id)))
		properties = append(properties, (C.cl_context_properties)(0))
		ctx, err := CreateContextFromTypeUnsafe(&properties[0], device_type.toCl(), nil, nil)
		return ctx, err
	}
	return nil, toError(C.CL_INVALID_DEVICE)
}

func (devType *DeviceType) toCl() C.cl_device_type {
	return C.cl_device_type(*devType)
}
