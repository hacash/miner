package cl

/*
#include "./opencl.h"
*/
import "C"

import "unsafe"

//////////////// Basic Types ////////////////
type SamplerAddressingMode int

const (
	SamplerAddressRepeat         SamplerAddressingMode = C.CL_ADDRESS_REPEAT
	SamplerAddressMirroredRepeat SamplerAddressingMode = C.CL_ADDRESS_MIRRORED_REPEAT
	SamplerAddressClampToEdge    SamplerAddressingMode = C.CL_ADDRESS_CLAMP_TO_EDGE
	SamplerAddressClamp          SamplerAddressingMode = C.CL_ADDRESS_CLAMP
	SamplerAddressNone           SamplerAddressingMode = C.CL_ADDRESS_NONE
)

type SamplerFilterMode int

const (
	SamplerFilterNearest SamplerFilterMode = C.CL_FILTER_NEAREST
	SamplerFilterLinear  SamplerFilterMode = C.CL_FILTER_LINEAR
)

//////////////// Abstract Types ////////////////
type Sampler struct {
	clSampler C.cl_sampler
}

//////////////// Basic Functions ////////////////
func releaseSampler(s *Sampler) {
	if s.clSampler != nil {
		C.clReleaseSampler(s.clSampler)
		s.clSampler = nil
	}
}

func retainSampler(s *Sampler) {
	if s.clSampler != nil {
		C.clRetainSampler(s.clSampler)
	}
}

//////////////// Abstract Functions ////////////////
func (s *Sampler) Release() {
	releaseSampler(s)
}

func (s *Sampler) Retain() {
	retainSampler(s)
}

func (s *Sampler) GetContext() (*Context, error) {
	if s.clSampler != nil {
		var outContext C.cl_context
		err := C.clGetSamplerInfo(s.clSampler, C.CL_SAMPLER_CONTEXT, C.size_t(unsafe.Sizeof(outContext)), unsafe.Pointer(&outContext), nil)
		return &Context{clContext: outContext, devices: nil}, toError(err)
	}
	return nil, toError(C.CL_INVALID_SAMPLER)
}

func (s *Sampler) GetReferenceCount() (int, error) {
	if s.clSampler != nil {
		var outCount C.cl_uint
		err := C.clGetSamplerInfo(s.clSampler, C.CL_SAMPLER_REFERENCE_COUNT, C.size_t(unsafe.Sizeof(outCount)), unsafe.Pointer(&outCount), nil)
		return int(outCount), toError(err)
	}
	return 0, toError(C.CL_INVALID_SAMPLER)
}

func (s *Sampler) GetNormalizedCoords() (bool, error) {
	if s.clSampler != nil {
		var outRes C.cl_bool
		err := C.clGetSamplerInfo(s.clSampler, C.CL_SAMPLER_NORMALIZED_COORDS, C.size_t(unsafe.Sizeof(outRes)), unsafe.Pointer(&outRes), nil)
		if toError(err) != nil {
			return false, toError(err)
		}
		switch {
		case outRes == C.CL_TRUE:
			return true, nil
		default:
			return false, nil
		}
	}
	return false, toError(C.CL_INVALID_SAMPLER)
}

func (s *Sampler) GetAddressingMode() (SamplerAddressingMode, error) {
	if s.clSampler != nil {
		var outRes C.cl_addressing_mode
		err := C.clGetSamplerInfo(s.clSampler, C.CL_SAMPLER_ADDRESSING_MODE, C.size_t(unsafe.Sizeof(outRes)), unsafe.Pointer(&outRes), nil)
		if toError(err) != nil {
			return -1, toError(err)
		}
		switch {
		case outRes == C.CL_ADDRESS_REPEAT:
			return SamplerAddressRepeat, nil
		case outRes == C.CL_ADDRESS_CLAMP_TO_EDGE:
			return SamplerAddressClampToEdge, nil
		case outRes == C.CL_ADDRESS_CLAMP:
			return SamplerAddressClamp, nil
		case outRes == C.CL_ADDRESS_NONE:
			return SamplerAddressNone, nil
		default:
			return -1, toError(C.CL_INVALID_SAMPLER)
		}
	}
	return -1, toError(C.CL_INVALID_SAMPLER)
}

func (s *Sampler) GetFilterMode() (SamplerFilterMode, error) {
	if s.clSampler != nil {
		var outRes C.cl_filter_mode
		err := C.clGetSamplerInfo(s.clSampler, C.CL_SAMPLER_FILTER_MODE, C.size_t(unsafe.Sizeof(outRes)), unsafe.Pointer(&outRes), nil)
		if toError(err) != nil {
			return -1, toError(err)
		}
		switch {
		case outRes == C.CL_FILTER_LINEAR:
			return SamplerFilterLinear, nil
		case outRes == C.CL_FILTER_NEAREST:
			return SamplerFilterNearest, nil
		default:
			return -1, toError(C.CL_INVALID_SAMPLER)
		}
	}
	return -1, toError(C.CL_INVALID_SAMPLER)
}

func (ctx *Context) CreateSampler(normalized_coors bool, addr_mode, filter_mode int) (*Sampler, error) {
	var err C.cl_int
	return &Sampler{C.clCreateSampler(ctx.clContext, clBool(normalized_coors), (C.cl_addressing_mode)(addr_mode), (C.cl_filter_mode)(filter_mode), &err)}, toError(err)
}
