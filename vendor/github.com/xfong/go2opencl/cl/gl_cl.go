//go:build gl || gl_dx || dx_gl
// +build gl gl_dx dx_gl

package cl

/*
#include <GL/gl.h>
#include <GL/glext.h>
#include "./opencl.h"

typedef CL_API_ENTRY cl_event (CL_API_CALL *clCreateEventFromGLsyncKHR_fn)(
    cl_context context,
    cl_GLsync  gl_sync,
    cl_int *   err);
static cl_int	(*clGetGLContextInfo) (const cl_context_properties *, cl_gl_context_info, size_t, void *, size_t *);
static cl_event	(*clCreateEventFromGLsync) (cl_context, cl_GLsync, cl_int *);

static void SetupGLSharing(cl_platform_id platform) {
	clGetGLContextInfoKHR_fn tmpPtr0 = NULL;
	tmpPtr0 = (clGetGLContextInfoKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clGetGLContextInfoKHR");
	clGetGLContextInfo = tmpPtr0;
	clCreateEventFromGLsyncKHR_fn tmpPtr1 = NULL;
	tmpPtr1 = (clCreateEventFromGLsyncKHR_fn)clGetExtensionFunctionAddressForPlatform(platform, "clCreateEventFromGLsyncKHR");
	clCreateEventFromGLsync = tmpPtr1;
}

static cl_int CLGetGLContextInfo(const cl_context_properties *properties, cl_gl_context_info gl_context_info, size_t param_value_size, void *param_value, size_t *param_value_size_ret) {
	return clGetGLContextInfo(properties, gl_context_info, param_value_size, param_value, param_value_size_ret);
}

static cl_event CLCreateEventFromGLsync(cl_context context, cl_GLsync glsync, cl_int * errCode) {
	return clCreateEventFromGLsync(context, glsync, errCode);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

//////////////// Basic Types ////////////////
type GLUint C.cl_GLuint
type GLInt C.cl_GLint
type GLEnum C.cl_GLenum
type GLSync C.cl_GLsync

var (
	ErrInvalidGlObject = errors.New("cl: Invalid Gl Object")
	ErrInvalidMipLevel = errors.New("cl: Invalid Mip Level")
)

const (
	ContextGL  ContextPropertiesId = C.CL_GL_CONTEXT_KHR
	ContextCGL ContextPropertiesId = C.CL_CGL_SHAREGROUP_KHR
	ContextEGL ContextPropertiesId = C.CL_EGL_DISPLAY_KHR
	ContextGLX ContextPropertiesId = C.CL_GLX_DISPLAY_KHR
	ContextWGL ContextPropertiesId = C.CL_WGL_HDC_KHR

	CommandAcquireGLObjects CommandType = C.CL_COMMAND_ACQUIRE_GL_OBJECTS
	CommandReleaseGLObjects CommandType = C.CL_COMMAND_RELEASE_GL_OBJECTS
)

type GLTargets int

const (
	GlTexture1D          GLTargets = C.GL_TEXTURE_1D
	GlTexture2D          GLTargets = C.GL_TEXTURE_2D
	GlTexture1DArray     GLTargets = C.GL_TEXTURE_1D_ARRAY
	GlTexture2DArray     GLTargets = C.GL_TEXTURE_2D_ARRAY
	GlTextureRect        GLTargets = C.GL_TEXTURE_RECTANGLE
	GlTexture3D          GLTargets = C.GL_TEXTURE_3D
	GlTextureBuffer      GLTargets = C.GL_TEXTURE_BUFFER
	GlTextureCubeMapPosX GLTargets = C.GL_TEXTURE_CUBE_MAP_POSITIVE_X
	GlTextureCubeMapNegX GLTargets = C.GL_TEXTURE_CUBE_MAP_NEGATIVE_X
	GlTextureCubeMapPosY GLTargets = C.GL_TEXTURE_CUBE_MAP_POSITIVE_Y
	GlTextureCubeMapNegY GLTargets = C.GL_TEXTURE_CUBE_MAP_NEGATIVE_Y
	GlTextureCubeMapPosZ GLTargets = C.GL_TEXTURE_CUBE_MAP_POSITIVE_Z
	GlTextureCubeMapNegZ GLTargets = C.GL_TEXTURE_CUBE_MAP_NEGATIVE_Z
)

type GLObjectTypes int

const (
	GlObjectBuffer         GLObjectTypes = C.CL_GL_OBJECT_BUFFER
	GlObjectTextureBuffer  GLObjectTypes = C.CL_GL_OBJECT_TEXTURE_BUFFER
	GlObjectTexture1D      GLObjectTypes = C.CL_GL_OBJECT_TEXTURE1D
	GlObjectTexture2D      GLObjectTypes = C.CL_GL_OBJECT_TEXTURE2D
	GlObjectTexture3D      GLObjectTypes = C.CL_GL_OBJECT_TEXTURE3D
	GlObjectTexture1DArray GLObjectTypes = C.CL_GL_OBJECT_TEXTURE1D_ARRAY
	GlObjectTexture2DArray GLObjectTypes = C.CL_GL_OBJECT_TEXTURE2D_ARRAY
	GlObjectRenderBuffer   GLObjectTypes = C.CL_GL_OBJECT_RENDERBUFFER
)

type GLTextureInfoPar int

const (
	GlTextureTarget GLTextureInfoPar = C.CL_GL_TEXTURE_TARGET
	GlMipmapLevel   GLTextureInfoPar = C.CL_GL_MIPMAP_LEVEL
)

type GLContextInfoType int

const (
	GLContextCurrentDevice GLContextInfoType = C.CL_CURRENT_DEVICE_FOR_GL_CONTEXT_KHR
	GLContextAllDevices    GLContextInfoType = C.CL_DEVICES_FOR_GL_CONTEXT_KHR
)

//////////////// Basic Functions ////////////////
func init() {
	errorMap[C.CL_INVALID_GL_OBJECT] = ErrInvalidGlObject
	errorMap[C.CL_INVALID_MIP_LEVEL] = ErrInvalidMipLevel

	gl_sharing_ext = true
	getGlCommandType = GlStatusToCommandType
}

func GlErrorCodeToCl(errCode int) C.cl_int {
	switch errCode {
	default:
		return C.CL_INVALID_GL_OBJECT
	case 0:
		return C.CL_SUCCESS
	case -6:
		return C.CL_OUT_OF_HOST_MEMORY
	case -30:
		return C.CL_INVALID_VALUE
	case -34:
		return C.CL_INVALID_CONTEXT
	case -60:
		return C.CL_INVALID_GL_OBJECT
	}
}

func GlStatusToCommandType(status C.cl_command_type) (bool, CommandType) {
	switch status {
	case C.CL_COMMAND_ACQUIRE_GL_OBJECTS:
		return true, CommandAcquireGLObjects
	case C.CL_COMMAND_RELEASE_GL_OBJECTS:
		return true, CommandReleaseGLObjects
	default:
		return false, -1
	}
}

func GlTargetToCl(targ GLTargets) C.cl_GLenum {
	return C.cl_GLenum(targ)
}

func GetCurrentDeviceInGlContext(go_ctx_properties []ContextPropertiesId) (*Device, error) {
	var c_ctx_properties []C.cl_context_properties
	defer C.free(c_ctx_properties)
	for i, prop := range go_ctx_properties {
		c_ctx_properties[i] = (C.cl_context_properties)(prop)
	}
	var device C.cl_device_id
	err := C.CLGetGLContextInfo(&c_ctx_properties[0], (C.cl_gl_context_info)(GLContextCurrentDevice), C.size_t(unsafe.Sizeof(device)), unsafe.Pointer(&device), nil)
	return &Device{id: device}, toError(err)
}

func GetAllDevicesInGlContext(go_ctx_properties []ContextPropertiesId) ([]*Device, error) {
	var c_ctx_properties []C.cl_context_properties
	defer C.free(c_ctx_properties)
	for i, prop := range go_ctx_properties {
		c_ctx_properties[i] = (C.cl_context_properties)(prop)
	}
	var devices []C.cl_device_id
	var devCnt C.size_t
	defer C.free(devices)
	err := C.CLGetGLContextInfo(&c_ctx_properties[0], C.CL_CURRENT_DEVICE_FOR_GL_CONTEXT_KHR, C.size_t(unsafe.Sizeof(devices)), unsafe.Pointer(&devices), &devCnt)
	devicesL := make([]C.cl_device_id, int(devCnt))
	defer C.free(devicesL)
	err = C.CLGetGLContextInfo(&c_ctx_properties[0], C.CL_CURRENT_DEVICE_FOR_GL_CONTEXT_KHR, devCnt, unsafe.Pointer(&devicesL), &devCnt)
	if toError(err) != nil {
		fmt.Printf("cl: failed to get all devices in GetAllDevicesInGlContext \n")
		return nil, toError(err)
	}
	DeviceList := make([]*Device, len(devices))
	for i, deviceId := range devicesL {
		DeviceList[i].id = deviceId
	}
	return DeviceList, nil
}

//////////////// Abstract Functions ////////////////
func (p *Platform) SetupGLSharing() {
	C.SetupGLSharing(p.id)
}

func (flag MemFlag) GlBufferCreateFlag() C.cl_mem_flags {
	switch flag {
	default:
		fmt.Printf("Unknown flag for CL/GL sharing")
		return C.CL_MEM_READ_WRITE
	case MemReadWrite:
		return C.CL_MEM_READ_WRITE
	case MemReadOnly:
		return C.CL_MEM_READ_ONLY
	case MemWriteOnly:
		return C.CL_MEM_WRITE_ONLY
	}
}

func (ctx *Context) CreateFromGlBuffer(flag MemFlag, GlBufferObject GLUint) (*MemObject, error) {
	var err C.int
	memobj := C.clCreateFromGLBuffer(ctx.clContext, flag.GlBufferCreateFlag(), (C.cl_GLuint)(GlBufferObject), &err)

	if toError(C.cl_int(err)) != nil {
		return nil, toError(GlErrorCodeToCl(int(err)))
	}
	GlBufferObj := &MemObject{clMem: memobj, size: 0}
	bufSize, sizeErr := GlBufferObj.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromGlBuffer \n")
		return nil, sizeErr
	}
	GlBufferObj.size = bufSize
	return GlBufferObj, nil
}

func (ctx *Context) CreateFromGlTexture(flag MemFlag, GlBufferObject GLUint, targ GLTargets, miplvl GLInt, texture GLUint) (*MemObject, error) {
	var err C.cl_int
	memobj := C.clCreateFromGLTexture(ctx.clContext, flag.GlBufferCreateFlag(), GlTargetToCl(targ), (C.cl_GLint)(miplvl), (C.cl_GLuint)(texture), &err)

	if toError(err) != nil {
		return nil, toError(err)
	}
	GlBufferObj := &MemObject{clMem: memobj, size: 0}
	bufSize, sizeErr := GlBufferObj.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromGlTexture2D \n")
		return nil, sizeErr
	}
	GlBufferObj.size = bufSize
	return GlBufferObj, nil
}

func (ctx *Context) CreateFromGlRenderBuffer(flag MemFlag, GlRenderBufferObject GLUint) (*MemObject, error) {
	var err C.cl_int
	memobj := C.clCreateFromGLRenderbuffer(ctx.clContext, flag.GlBufferCreateFlag(), (C.cl_GLuint)(GlRenderBufferObject), &err)

	if toError(err) != nil {
		return nil, toError(GlErrorCodeToCl(int(err)))
	}
	GlBufferObj := &MemObject{clMem: memobj, size: 0}
	bufSize, sizeErr := GlBufferObj.GetSize()
	if sizeErr != nil {
		fmt.Printf("Unable to get buffer size in CreateFromGlRenderBuffer \n")
		return nil, sizeErr
	}
	GlBufferObj.size = bufSize
	return GlBufferObj, nil
}

func (ctx *Context) CreateEventFromGLsync(gl_sync GLSync) (*Event, error) {
	var errCode C.cl_int
	defer C.free(errCode)
	event := C.CLCreateEventFromGLsync(ctx.clContext, (C.cl_GLsync)(gl_sync), &errCode)
	return newEvent(event), toError(errCode)
}

func (mb *MemObject) GetGlObjectInfo() (GLObjectTypes, *GLUint) {
	var glObjType C.cl_gl_object_type
	var globj_name C.cl_GLuint
	err := C.clGetGLObjectInfo(mb.clMem, &glObjType, &globj_name)
	if toError(err) != nil {
		fmt.Printf("cl: failed to get GL object info \n")
		return -1, nil
	}
	name := (GLUint)(globj_name)
	return (GLObjectTypes)(glObjType), &name
}

func (mb *MemObject) GetTextureTarget() (GLTargets, error) {
	var val C.cl_GLenum
	err := C.clGetGLTextureInfo(mb.clMem, (C.cl_gl_texture_info)(GlTextureTarget), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
	return (GLTargets)(val), toError(err)
}

func (mb *MemObject) GetTextureMipmapLevel() (GLInt, error) {
	var val C.cl_GLint
	err := C.clGetGLTextureInfo(mb.clMem, (C.cl_gl_texture_info)(GlMipmapLevel), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
	return (GLInt)(val), toError(err)
}

func (q *CommandQueue) EnqueueAcquireGlObjects(memObjList []*MemObject, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	memObjCnt := len(memObjList)
	memObjs := make([]C.cl_mem, memObjCnt)
	for i, mPtr := range memObjList {
		memObjs[i] = mPtr.clMem
	}
	err := C.clEnqueueAcquireGLObjects(q.clQueue, (C.cl_uint)(memObjCnt), &memObjs[0], (C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}

func (q *CommandQueue) EnqueueReleaseGlObjects(memObjList []*MemObject, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	memObjCnt := len(memObjList)
	memObjs := make([]C.cl_mem, memObjCnt)
	for i, mPtr := range memObjList {
		memObjs[i] = mPtr.clMem
	}
	err := C.clEnqueueReleaseGLObjects(q.clQueue, (C.cl_uint)(memObjCnt), &memObjs[0], (C.cl_uint)(len(eventWaitList)), eventListPtr(eventWaitList), &event)
	return newEvent(event), toError(err)
}
