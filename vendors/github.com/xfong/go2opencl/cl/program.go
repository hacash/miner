package cl

/*
#include "./opencl.h"
extern void go_program_notify(cl_program alt_program, void *user_data);
extern void go_compile_program_notify(cl_program alt_program, void *user_data);
extern void go_link_program_notify(cl_program alt_program, void *user_data);
static void CL_CALLBACK c_program_notify(cl_program alt_program, void *user_data) {
        go_program_notify((cl_program) alt_program, (void *)user_data);
}

static void CL_CALLBACK c_compile_program_notify(cl_program alt_program, void *user_data) {
        go_compile_program_notify((cl_program) alt_program, (void *)user_data);
}

static void CL_CALLBACK c_link_program_notify(cl_program alt_program, void *user_data) {
        go_link_program_notify((cl_program) alt_program, (void *)user_data);
}

static cl_int CLBuildProgram(      			cl_program 				program,
                                                        cl_uint                                 num_devices,
                                                  const cl_device_id *                    devices,
						  const char *				build_options,
                                                        void *                                  user_data) {
        return clBuildProgram(program, num_devices, devices, build_options, c_program_notify, user_data);
}

static cl_int CLCompileProgram(                           cl_program                              program,
                                                        cl_uint                                 num_devices,
                                                  const cl_device_id *                    devices,
                                                  const char *                          build_options,
							cl_uint				num_headers,
						const cl_program *			headers,
						const char **				header_names,
                                                        void *                                  user_data) {
        return clCompileProgram(program, num_devices, devices, build_options, num_headers, headers, header_names, c_compile_program_notify, user_data);
}

static cl_program CLLinkProgram(                           cl_context                              context,
                                                        cl_uint                                 num_devices,
                                                  const cl_device_id *                    devices,
                                                  const char *                          build_options,
							cl_uint				num_programs,
						const cl_program *			in_programs,
                                                        void *                                  user_data,
							cl_int * err_ret) {
        return clLinkProgram(context, num_devices, devices, build_options, num_programs, in_programs, c_link_program_notify, user_data, err_ret);
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

// ////////////// Basic Types ////////////////
type BuildStatus int

const (
	BuildStatusSuccess    BuildStatus = C.CL_BUILD_SUCCESS
	BuildStatusNone       BuildStatus = C.CL_BUILD_NONE
	BuildStatusError      BuildStatus = C.CL_BUILD_ERROR
	BuildStatusInProgress BuildStatus = C.CL_BUILD_IN_PROGRESS
)

type ProgramBinaryTypes int

const (
	ProgramBinaryTypeNone           ProgramBinaryTypes = C.CL_PROGRAM_BINARY_TYPE_NONE
	ProgramBinaryTypeCompiledObject ProgramBinaryTypes = C.CL_PROGRAM_BINARY_TYPE_COMPILED_OBJECT
	ProgramBinaryTypeLibrary        ProgramBinaryTypes = C.CL_PROGRAM_BINARY_TYPE_LIBRARY
	ProgramBinaryTypeExecutable     ProgramBinaryTypes = C.CL_PROGRAM_BINARY_TYPE_EXECUTABLE
)

// ////////////// Abstract Types ////////////////
type BuildError struct {
	Message string
	Device  *Device
}

func (e BuildError) Error() string {
	if e.Device != nil {
		return fmt.Sprintf("cl: build error on %q: %s", e.Device.Name(), e.Message)
	} else {
		return fmt.Sprintf("cl: build error: %s", e.Message)
	}
}

type Program struct {
	clProgram C.cl_program
	devices   []*Device
}

type ProgramHeaders struct {
	codes Program
	names string
}

// //////////////// Supporting Types ////////////////
type CL_program_notify func(alt_program C.cl_program, user_data unsafe.Pointer)

var program_notify map[unsafe.Pointer]CL_program_notify

type CL_compile_program_notify func(alt_program C.cl_program, user_data unsafe.Pointer)

var compile_program_notify map[unsafe.Pointer]CL_compile_program_notify

type CL_link_program_notify func(alt_program C.cl_program, user_data unsafe.Pointer)

var link_program_notify map[unsafe.Pointer]CL_link_program_notify

// //////////////// Basic Functions ////////////////
func init() {
	program_notify = make(map[unsafe.Pointer]CL_program_notify)
	compile_program_notify = make(map[unsafe.Pointer]CL_compile_program_notify)
	link_program_notify = make(map[unsafe.Pointer]CL_link_program_notify)
}

//export go_program_notify
func go_program_notify(alt_program C.cl_program, user_data unsafe.Pointer) {
	var c_user_data []unsafe.Pointer
	c_user_data = *(*[]unsafe.Pointer)(user_data)
	program_notify[c_user_data[1]](alt_program, c_user_data[0])
}

//export go_compile_program_notify
func go_compile_program_notify(alt_program C.cl_program, user_data unsafe.Pointer) {
	var c_user_data []unsafe.Pointer
	c_user_data = *(*[]unsafe.Pointer)(user_data)
	compile_program_notify[c_user_data[1]](alt_program, c_user_data[0])
}

//export go_link_program_notify
func go_link_program_notify(alt_program C.cl_program, user_data unsafe.Pointer) {
	var c_user_data []unsafe.Pointer
	c_user_data = *(*[]unsafe.Pointer)(user_data)
	link_program_notify[c_user_data[1]](alt_program, c_user_data[0])
}

// ////////////// Basic Functions ////////////////
func releaseProgram(p *Program) {
	if p.clProgram != nil {
		C.clReleaseProgram(p.clProgram)
		p.clProgram = nil
	}
}

func retainProgram(p *Program) {
	if p.clProgram != nil {
		C.clRetainProgram(p.clProgram)
	}
}

// ////////////// Abstract Functions ////////////////
func (p *Program) Release() {
	releaseProgram(p)
}

func (p *Program) Retain() {
	retainProgram(p)
}

func (p *Program) BuildProgram(devices []*Device, options string) error {
	var optBuffer bytes.Buffer
	optBuffer.WriteString("-cl-std=CL1.2 -cl-kernel-arg-info ")
	var cOptions *C.char
	if options != "" {
		optBuffer.WriteString(options)
	}
	cOptions = C.CString(optBuffer.String())
	defer C.free(unsafe.Pointer(cOptions))

	var deviceList []C.cl_device_id
	var deviceListPtr *C.cl_device_id
	numDevices := C.cl_uint(len(devices))
	if devices != nil && len(devices) > 0 {
		deviceList = buildDeviceIdList(devices)
		deviceListPtr = &deviceList[0]
	}
	if err := C.clBuildProgram(p.clProgram, numDevices, deviceListPtr, cOptions, nil, nil); err != C.CL_SUCCESS {
		buffer := make([]byte, 4096)
		var bLen C.size_t
		var err C.cl_int

		for _, dev := range p.devices {
			for i := 2; i >= 0; i-- {
				err = C.clGetProgramBuildInfo(p.clProgram, dev.id, C.CL_PROGRAM_BUILD_LOG, C.size_t(len(buffer)), unsafe.Pointer(&buffer[0]), &bLen)
				if err == C.CL_INVALID_VALUE && i > 0 && bLen < 1024*1024 {
					// INVALID_VALUE probably means our buffer isn't large enough
					buffer = make([]byte, bLen)
				} else {
					break
				}
			}
			if err != C.CL_SUCCESS {
				return toError(err)
			}

			if bLen > 1 {
				return BuildError{
					Device:  dev,
					Message: string(buffer[:bLen-1]),
				}
			}
		}

		return BuildError{
			Device:  nil,
			Message: "build failed and produced no log entries",
		}
	}
	return nil
}

func (p *Program) CreateKernel(name string) (*Kernel, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	var err C.cl_int
	clKernel := C.clCreateKernel(p.clProgram, cName, &err)
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	kernel := &Kernel{clKernel: clKernel, name: name}
	runtime.SetFinalizer(kernel, releaseKernel)
	return kernel, nil
}

func (ctx *Context) CreateProgramWithSource(sources []string) (*Program, error) {
	cSources := make([]*C.char, len(sources))
	for i, s := range sources {
		cs := C.CString(s)
		cSources[i] = cs
		defer C.free(unsafe.Pointer(cs))
	}
	var err C.cl_int
	clProgram := C.clCreateProgramWithSource(ctx.clContext, C.cl_uint(len(sources)), &cSources[0], nil, &err)
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	if clProgram == nil {
		return nil, ErrUnknown
	}
	program := &Program{clProgram: clProgram, devices: ctx.devices}
	runtime.SetFinalizer(program, releaseProgram)
	return program, nil
}

func (ctx *Context) CreateProgramWithBuiltInKernels(devices []*Device, kernel_names []string) (*Program, error) {
	cSources := make([]*C.char, 1)
	merge_string := strings.Join(kernel_names, ";")
	cs := C.CString(merge_string)
	cSources[0] = cs
	defer C.free(unsafe.Pointer(cs))
	var deviceList []C.cl_device_id
	var deviceListPtr *C.cl_device_id
	numDevices := C.cl_uint(len(devices))
	if devices != nil && len(devices) > 0 {
		deviceList = buildDeviceIdList(devices)
		deviceListPtr = &deviceList[0]
	}
	var err C.cl_int
	clProgram := C.clCreateProgramWithBuiltInKernels(ctx.clContext, numDevices, deviceListPtr, cSources[0], &err)
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	if clProgram == nil {
		return nil, ErrUnknown
	}
	program := &Program{clProgram: clProgram, devices: ctx.devices}
	runtime.SetFinalizer(program, releaseProgram)
	return program, nil
}

func (p *Program) CompileProgram(devices []*Device, options string, program_headers []*ProgramHeaders) error {
	var cOptions *C.char
	if options != "" {
		cOptions = C.CString(options)
		defer C.free(unsafe.Pointer(cOptions))
	}
	var deviceList []C.cl_device_id
	var deviceListPtr *C.cl_device_id
	numDevices := C.cl_uint(len(devices))
	if devices != nil && len(devices) > 0 {
		deviceList = buildDeviceIdList(devices)
		deviceListPtr = &deviceList[0]
	}
	num_headers := len(program_headers)
	cHeaders := make([]C.cl_program, num_headers)
	cHeader_names := make([]*C.char, num_headers)
	for idx, ph := range program_headers {
		chs := ph.codes
		chn := C.CString(ph.names)
		cHeaders[idx] = chs.clProgram
		cHeader_names[idx] = chn
		defer C.free(unsafe.Pointer(&chs))
		defer C.free(unsafe.Pointer(&chn))
	}
	err := C.clCompileProgram(p.clProgram, numDevices, deviceListPtr, cOptions, C.cl_uint(num_headers), &cHeaders[0], &cHeader_names[0], nil, nil)
	if err != C.CL_SUCCESS {
		buffer := make([]byte, 4096)
		var bLen C.size_t
		var err C.cl_int

		for _, dev := range p.devices {
			for i := 2; i >= 0; i-- {
				err = C.clGetProgramBuildInfo(p.clProgram, dev.id, C.CL_PROGRAM_BUILD_LOG, C.size_t(len(buffer)), unsafe.Pointer(&buffer[0]), &bLen)
				if err == C.CL_INVALID_VALUE && i > 0 && bLen < 1024*1024 {
					// INVALID_VALUE probably means our buffer isn't large enough
					buffer = make([]byte, bLen)
				} else {
					break
				}
			}
			if err != C.CL_SUCCESS {
				return toError(err)
			}

			if bLen > 1 {
				return BuildError{
					Device:  dev,
					Message: string(buffer[:bLen-1]),
				}
			}
		}

		return BuildError{
			Device:  nil,
			Message: "build failed and produced no log entries",
		}
	}
	return nil
}

func (p *Program) CompileProgramWithCallback(devices []*Device, options string, program_headers []*ProgramHeaders, user_data unsafe.Pointer) error {
	var cOptions *C.char
	if options != "" {
		cOptions = C.CString(options)
		defer C.free(unsafe.Pointer(cOptions))
	}
	var deviceList []C.cl_device_id
	var deviceListPtr *C.cl_device_id
	numDevices := C.cl_uint(len(devices))
	if devices != nil && len(devices) > 0 {
		deviceList = buildDeviceIdList(devices)
		deviceListPtr = &deviceList[0]
	}
	num_headers := len(program_headers)
	cHeaders := make([]C.cl_program, num_headers)
	cHeader_names := make([]*C.char, num_headers)
	for idx, ph := range program_headers {
		chs := ph.codes
		chn := C.CString(ph.names)
		cHeaders[idx] = chs.clProgram
		cHeader_names[idx] = chn
		defer C.free(unsafe.Pointer(&chs))
		defer C.free(unsafe.Pointer(&chn))
	}
	err := C.CLCompileProgram(p.clProgram, numDevices, deviceListPtr, cOptions, C.cl_uint(num_headers), &cHeaders[0], &cHeader_names[0], user_data)
	if err != C.CL_SUCCESS {
		buffer := make([]byte, 4096)
		var bLen C.size_t
		var err C.cl_int

		for _, dev := range p.devices {
			for i := 2; i >= 0; i-- {
				err = C.clGetProgramBuildInfo(p.clProgram, dev.id, C.CL_PROGRAM_BUILD_LOG, C.size_t(len(buffer)), unsafe.Pointer(&buffer[0]), &bLen)
				if err == C.CL_INVALID_VALUE && i > 0 && bLen < 1024*1024 {
					// INVALID_VALUE probably means our buffer isn't large enough
					buffer = make([]byte, bLen)
				} else {
					break
				}
			}
			if err != C.CL_SUCCESS {
				return toError(err)
			}

			if bLen > 1 {
				return BuildError{
					Device:  dev,
					Message: string(buffer[:bLen-1]),
				}
			}
		}

		return BuildError{
			Device:  nil,
			Message: "build failed and produced no log entries",
		}
	}
	return nil
}

func (ctx *Context) LinkProgram(programs []*Program, devices []*Device, options string) (*Program, error) {
	var cOptions *C.char
	if options != "" {
		cOptions = C.CString(options)
		defer C.free(unsafe.Pointer(cOptions))
	}
	var deviceList []C.cl_device_id
	var deviceListPtr *C.cl_device_id
	numDevices := C.cl_uint(len(devices))
	if devices != nil && len(devices) > 0 {
		deviceList = buildDeviceIdList(devices)
		deviceListPtr = &deviceList[0]
	}
	programList := make([]C.cl_program, len(programs))
	for idx, progId := range programs {
		programList[idx] = progId.clProgram
	}
	var err C.cl_int
	programExe := C.clLinkProgram(ctx.clContext, numDevices, deviceListPtr, cOptions, C.cl_uint(len(programs)), &programList[0], nil, nil, &err)
	p := &Program{clProgram: programExe, devices: devices}
	if err != C.CL_SUCCESS {
		buffer := make([]byte, 4096)
		var bLen C.size_t
		var err C.cl_int

		for _, dev := range p.devices {
			for i := 2; i >= 0; i-- {
				err = C.clGetProgramBuildInfo(p.clProgram, dev.id, C.CL_PROGRAM_BUILD_LOG, C.size_t(len(buffer)), unsafe.Pointer(&buffer[0]), &bLen)
				if err == C.CL_INVALID_VALUE && i > 0 && bLen < 1024*1024 {
					// INVALID_VALUE probably means our buffer isn't large enough
					buffer = make([]byte, bLen)
				} else {
					break
				}
			}
			if err != C.CL_SUCCESS {
				return nil, toError(err)
			}

			if bLen > 1 {
				return nil, BuildError{
					Device:  dev,
					Message: string(buffer[:bLen-1]),
				}
			}
		}

		return nil, BuildError{
			Device:  nil,
			Message: "build failed and produced no log entries",
		}
	}
	return p, nil
}

func (ctx *Context) LinkProgramWithCallback(programs []*Program, devices []*Device, options string, user_data unsafe.Pointer) (*Program, error) {
	var cOptions *C.char
	if options != "" {
		cOptions = C.CString(options)
		defer C.free(unsafe.Pointer(cOptions))
	}
	var deviceList []C.cl_device_id
	var deviceListPtr *C.cl_device_id
	numDevices := C.cl_uint(len(devices))
	if devices != nil && len(devices) > 0 {
		deviceList = buildDeviceIdList(devices)
		deviceListPtr = &deviceList[0]
	}
	programList := make([]C.cl_program, len(programs))
	for idx, progId := range programs {
		programList[idx] = progId.clProgram
	}
	var err C.cl_int
	programExe := C.CLLinkProgram(ctx.clContext, numDevices, deviceListPtr, cOptions, C.cl_uint(len(programs)), &programList[0], user_data, &err)
	p := &Program{clProgram: programExe, devices: devices}
	if err != C.CL_SUCCESS {
		buffer := make([]byte, 4096)
		var bLen C.size_t
		var err C.cl_int

		for _, dev := range p.devices {
			for i := 2; i >= 0; i-- {
				err = C.clGetProgramBuildInfo(p.clProgram, dev.id, C.CL_PROGRAM_BUILD_LOG, C.size_t(len(buffer)), unsafe.Pointer(&buffer[0]), &bLen)
				if err == C.CL_INVALID_VALUE && i > 0 && bLen < 1024*1024 {
					// INVALID_VALUE probably means our buffer isn't large enough
					buffer = make([]byte, bLen)
				} else {
					break
				}
			}
			if err != C.CL_SUCCESS {
				return nil, toError(err)
			}

			if bLen > 1 {
				return nil, BuildError{
					Device:  dev,
					Message: string(buffer[:bLen-1]),
				}
			}
		}

		return nil, BuildError{
			Device:  nil,
			Message: "build failed and produced no log entries",
		}
	}
	return p, nil
}

func (p *Program) GetBuildStatus(device *Device) (BuildStatus, error) {
	var buildStatus C.cl_build_status
	err := C.clGetProgramBuildInfo(p.clProgram, device.id, C.CL_PROGRAM_BUILD_STATUS, C.size_t(unsafe.Sizeof(buildStatus)), unsafe.Pointer(&buildStatus), nil)
	return BuildStatus(buildStatus), toError(err)
}

func (p *Program) GetBuildOptions(device *Device) (string, error) {
	var strC [1024]C.char
	var strN C.size_t
	if err := C.clGetProgramBuildInfo(p.clProgram, device.id, C.CL_PROGRAM_BUILD_OPTIONS, 1024, unsafe.Pointer(&strC), &strN); err != C.CL_SUCCESS {
		panic("Should never fail")
		return "", toError(err)
	}

	// OpenCL strings are NUL-terminated, and the terminator is included in strN
	// Go strings aren't NUL-terminated, so subtract 1 from the length
	return C.GoStringN((*C.char)(unsafe.Pointer(&strC)), C.int(strN-1)), nil
}

func (p *Program) GetBuildLog(device *Device) (string, error) {
	var strC [1024]C.char
	var strN C.size_t
	if err := C.clGetProgramBuildInfo(p.clProgram, device.id, C.CL_PROGRAM_BUILD_LOG, 1024, unsafe.Pointer(&strC), &strN); err != C.CL_SUCCESS {
		panic("Should never fail")
		return "", toError(err)
	}

	// OpenCL strings are NUL-terminated, and the terminator is included in strN
	// Go strings aren't NUL-terminated, so subtract 1 from the length
	return C.GoStringN((*C.char)(unsafe.Pointer(&strC)), C.int(strN-1)), nil
}

func (p *Program) GetProgramBinaryType(device *Device) (ProgramBinaryTypes, error) {
	var binType C.cl_program_binary_type
	err := C.clGetProgramBuildInfo(p.clProgram, device.id, C.CL_PROGRAM_BINARY_TYPE, C.size_t(unsafe.Sizeof(binType)), unsafe.Pointer(&binType), nil)
	return ProgramBinaryTypes(binType), toError(err)
}

func (p *Program) GetReferenceCount() (int, error) {
	var val C.cl_uint
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_REFERENCE_COUNT, C.size_t(unsafe.Sizeof(val)), (unsafe.Pointer)(&val), nil); err != C.CL_SUCCESS {
		panic("Should never fail")
		return -1, toError(err)
	}

	return int(val), nil
}

func (p *Program) GetContext() (*Context, error) {
	var val C.cl_context
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_CONTEXT, C.size_t(unsafe.Sizeof(val)), (unsafe.Pointer)(&val), nil); err != C.CL_SUCCESS {
		panic("Should never fail")
		return nil, toError(err)
	}

	return &Context{clContext: val, devices: nil}, nil
}

func (p *Program) GetDeviceCount() (int, error) {
	var val C.cl_uint
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_NUM_DEVICES, C.size_t(unsafe.Sizeof(val)), (unsafe.Pointer)(&val), nil); err != C.CL_SUCCESS {
		panic("Should never fail")
		return -1, toError(err)
	}

	return int(val), nil
}

func (p *Program) GetDevices() ([]*Device, error) {
	var val C.cl_device_id
	var arr []C.cl_device_id
	var cnts C.size_t
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_DEVICES, C.size_t(unsafe.Sizeof(val)), (unsafe.Pointer)(&arr), &cnts); err != C.CL_SUCCESS {
		panic("Should never fail")
		return nil, toError(err)
	}

	returnDevices := make([]*Device, int(cnts))
	for i := 0; i < int(cnts); i++ {
		returnDevices[i] = &Device{id: arr[i]}
	}
	return returnDevices, nil
}

func (p *Program) GetSource() (string, error) {
	var strC [1024]C.char
	var strN C.size_t
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_SOURCE, 1024, unsafe.Pointer(&strC), &strN); err != C.CL_SUCCESS {
		panic("Should never fail")
		return "", toError(err)
	}

	// OpenCL strings are NUL-terminated, and the terminator is included in strN
	// Go strings aren't NUL-terminated, so subtract 1 from the length
	return C.GoStringN((*C.char)(unsafe.Pointer(&strC)), C.int(strN-1)), nil
}

func (p *Program) GetBinarySizes() ([]int, error) {
	var val C.size_t
	var val2 C.size_t
	var arr []C.size_t
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_BINARY_SIZES, C.size_t(unsafe.Sizeof(val)), (unsafe.Pointer)(&arr), &val2); err != C.CL_SUCCESS {
		panic("Should never fail")
		return nil, toError(err)
	}
	fmt.Println(arr, val, val2)
	returnCount := make([]int, int(val))
	for i := 0; i < int(val); i++ {
		returnCount[i] = int(arr[i])
	}
	return returnCount, nil
}

func (p *Program) GetBinaries() ([]*uint8, error) {
	var item *uint8
	var val C.size_t
	var arr []*uint8
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_BINARIES, C.size_t(unsafe.Sizeof(item)), (unsafe.Pointer)(&arr), &val); err != C.CL_SUCCESS {
		panic("Should never fail")
		return nil, toError(err)
	}

	returnBinaries := make([]*uint8, int(val))
	for i := 0; i < int(val); i++ {
		returnBinaries[i] = arr[i]
	}
	return returnBinaries, nil
}

func (p *Program) GetKernelCounts() (int, error) {
	var val C.size_t
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_NUM_KERNELS, C.size_t(unsafe.Sizeof(val)), (unsafe.Pointer)(&val), nil); err != C.CL_SUCCESS {
		panic("Should never fail")
		return -1, toError(err)
	}
	return int(val), nil
}

func (p *Program) GetKernelNames() (string, error) {
	var strC [1024]C.char
	var strN C.size_t
	if err := C.clGetProgramInfo(p.clProgram, C.CL_PROGRAM_KERNEL_NAMES, 1024, unsafe.Pointer(&strC), &strN); err != C.CL_SUCCESS {
		panic("Should never fail")
		return "", toError(err)
	}

	// OpenCL strings are NUL-terminated, and the terminator is included in strN
	// Go strings aren't NUL-terminated, so subtract 1 from the length
	return C.GoStringN((*C.char)(unsafe.Pointer(&strC)), C.int(strN-1)), nil
}

func (ctx *Context) CreateProgramWithBinary(deviceList []*Device, program_lengths []int, program_binaries []*uint8) (*Program, error) {
	var binary_in []*C.uchar
	device_list_in := make([]C.cl_device_id, len(deviceList))
	binary_lengths := make([]C.size_t, len(program_lengths))
	defer C.free(unsafe.Pointer(&binary_in))
	defer C.free(unsafe.Pointer(&binary_lengths))
	defer C.free(unsafe.Pointer(&device_list_in))
	var binErr []C.cl_int
	var err C.cl_int
	for i, bin_val := range program_binaries {
		binary_lengths[i] = C.size_t(program_lengths[i])
		binary_in[i] = (*C.uchar)(bin_val)
	}
	for i, devItem := range deviceList {
		device_list_in[i] = devItem.id
	}
	clProgram := C.clCreateProgramWithBinary(ctx.clContext, C.cl_uint(len(deviceList)), &device_list_in[0], &binary_lengths[0], &binary_in[0], &binErr[0], &err)
	for i := range binary_lengths {
		if binErr[i] != C.CL_SUCCESS {
			errMsg := int(binErr[i])
			switch errMsg {
			default:
				fmt.Printf("Unknown error when loading binary %d \n", i)
			case C.CL_INVALID_VALUE:
				fmt.Printf("Loading empty binary %d \n", i)
			case C.CL_INVALID_BINARY:
				fmt.Printf("Loading an invalid binary %d \n", i)
			}
		}
	}
	if err != C.CL_SUCCESS {
		return nil, toError(err)
	}
	if clProgram == nil {
		return nil, ErrUnknown
	}
	program := &Program{clProgram: clProgram, devices: ctx.devices}
	runtime.SetFinalizer(program, releaseProgram)
	return program, nil
}

func (pf *Platform) UnloadCompiler() error {
	return toError(C.clUnloadPlatformCompiler(pf.id))
}
