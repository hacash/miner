package cl

/*
#include "cl.h"
*/
import "C"

import (
	"bytes"
	"fmt"
	"unsafe"
)

type Platform struct {
	id   CL_platform_id
	name string
}

func (p *Platform) Name() string {
	return p.name
}

type Device struct {
	id   CL_device_id
	name string
}

type Context struct {
	clContext CL_context
	devices   []*Device
}

type Program struct {
	clProgram CL_program
	devices   []*Device
}

type Kernel struct {
	clKernel CL_kernel
	name     string
}

type CommandQueue struct {
	clQueue CL_command_queue
	device  *Device
}

type MemObject struct {
	clMem CL_mem
	size  int
}

type Event struct {
	clEvent CL_event
}

func (m *MemObject) Retain() error {
	ret := CLRetainMemObject(m.clMem)
	if ret != CL_SUCCESS {
		return fmt.Errorf("CLRetainMemObject Error Code = %d", ret)
	}
	return nil
}

func (q *CommandQueue) EnqueueWriteBufferByte(buffer *MemObject, blocking bool, offset int, data []byte) (*Event, error) {
	dataPtr := unsafe.Pointer(&data[0])
	dataSize := int(unsafe.Sizeof(data[0])) * len(data)
	var evt CL_event
	var is_blocking int = 0
	if blocking {
		is_blocking = 1
	}
	ret := CLEnqueueWriteBuffer(q.clQueue, buffer.clMem, CL_bool(is_blocking), CL_size_t(offset), CL_size_t(dataSize), dataPtr, CL_uint(0), nil, &evt)
	if ret != CL_SUCCESS {
		return nil, fmt.Errorf("CLEnqueueWriteBuffer Error Code = %d", ret)
	}
	return &Event{clEvent: evt}, nil
}

func (q *CommandQueue) EnqueueReadBufferByte(buffer *MemObject, blocking bool, offset int, data []byte) (*Event, error) {
	dataPtr := unsafe.Pointer(&data[0])
	dataSize := int(unsafe.Sizeof(data[0])) * len(data)
	var is_blocking int = 0
	if blocking {
		is_blocking = 1
	}
	var evt CL_event
	ret := CLEnqueueReadBuffer(q.clQueue, buffer.clMem, CL_bool(is_blocking), CL_size_t(offset), CL_size_t(dataSize), dataPtr, CL_uint(0), nil, &evt)
	if ret != CL_SUCCESS {
		return nil, fmt.Errorf("CLEnqueueReadBuffer Error Code = %d", ret)
	}
	return &Event{clEvent: evt}, nil
}

func (q *CommandQueue) EnqueueNDRangeKernel(kernel *Kernel, globalWorkOffset, globalWorkSize, localWorkSize []int, eventWaitList []*Event) (*Event, error) {

	var evt CL_event
	//var evts = make([]CL_event, 0)
	// 1
	var szs1 = make([]CL_size_t, len(globalWorkOffset))
	for i := 0; i < len(globalWorkOffset); i++ {
		szs1[i] = CL_size_t(globalWorkOffset[i])
	}
	// 2
	var szs2 = make([]CL_size_t, len(globalWorkSize))
	for i := 0; i < len(globalWorkSize); i++ {
		szs2[i] = CL_size_t(globalWorkSize[i])
	}
	// 3
	var szs3 = make([]CL_size_t, len(localWorkSize))
	for i := 0; i < len(localWorkSize); i++ {
		szs3[i] = CL_size_t(localWorkSize[i])
	}
	ret := CLEnqueueNDRangeKernel(q.clQueue, kernel.clKernel, CL_uint(len(globalWorkSize)), szs1, szs2, szs3, CL_uint(0), nil, &evt)
	if ret != CL_SUCCESS {
		return nil, fmt.Errorf("CLEnqueueReadBuffer Error Code = %d", ret)
	}
	return &Event{clEvent: evt}, nil
}

func (q *CommandQueue) Finish() error {
	ret := CLFinish(q.clQueue)
	if ret != CL_SUCCESS {
		return fmt.Errorf("CLFinish Error Code = %d", ret)
	}
	return nil
}

func (q *CommandQueue) Retain() error {
	ret := CLRetainCommandQueue(q.clQueue)
	if ret != CL_SUCCESS {
		return fmt.Errorf("CLRetainCommandQueue Error Code = %d", ret)
	}
	return nil
}

func (p *Device) Name() string {
	return p.name
}

func (p *Device) MaxWorkGroupSize() int {
	var retsize CL_size_t
	CLGetDeviceInfo(p.id, CL_DEVICE_MAX_WORK_GROUP_SIZE, 0, nil, &retsize)
	var val interface{}
	val = CL_size_t(0)
	CLGetDeviceInfo(p.id, CL_DEVICE_MAX_WORK_GROUP_SIZE, retsize, &val, nil)
	return int(val.(CL_size_t))
}

func (k *Kernel) SetArgs(args ...interface{}) error {
	var ret CL_int
	for index, arg := range args {
		var argsize = CL_size_t(unsafe.Sizeof(arg))
		var argptr unsafe.Pointer = unsafe.Pointer(&arg)
		// number
		switch arg.(type) {
		case uint8:
			argsize = CL_size_t(1)
			argv := arg.(uint8)
			argptr = unsafe.Pointer(&argv)
		case int8:
			argsize = CL_size_t(1)
			argv := arg.(int8)
			argptr = unsafe.Pointer(&argv)
		case uint32:
			argsize = CL_size_t(4)
			argv := arg.(uint32)
			argptr = unsafe.Pointer(&argv)
		case int32:
			argsize = CL_size_t(4)
			argv := arg.(int32)
			argptr = unsafe.Pointer(&argv)
		case uint64:
			argsize = CL_size_t(8)
			argv := arg.(uint64)
			argptr = unsafe.Pointer(&argv)
		case int64:
			argsize = CL_size_t(8)
			argv := arg.(int64)
			argptr = unsafe.Pointer(&argv)
		case float32:
			argsize = CL_size_t(4)
			argv := arg.(float32)
			argptr = unsafe.Pointer(&argv)
		case float64:
			argsize = CL_size_t(8)
			argv := arg.(float64)
			argptr = unsafe.Pointer(&argv)
		}
		// *MemObject
		if mem, ok := arg.(*MemObject); ok {
			argsize = CL_size_t(unsafe.Sizeof(mem.clMem.cl_mem))
			argptr = unsafe.Pointer(&mem.clMem.cl_mem)
		}
		ret = CLSetKernelArg(k.clKernel, CL_uint(index), argsize, argptr)
		if ret != CL_SUCCESS {
			return fmt.Errorf("arg index = %d, argsize=%d CLSetKernelArg Error Code = %d", index, argsize, ret)
		}
	}
	return nil
}

func (k *Kernel) Retain() error {
	ret := CLRetainKernel(k.clKernel)
	if ret != CL_SUCCESS {
		return fmt.Errorf("CLRetainKernel Error Code = %d", ret)
	}
	return nil
}

func (c *Context) CreateProgramWithBinary(device []*Device, binbts [][]byte) (*Program, error) {
	if len(device) != len(binbts) {
		return nil, fmt.Errorf("device length must = binbts length")
	}
	var err CL_int = 0
	var num = CL_uint(len(device))
	var dids = getDeviceIds(device)
	var lens = make([]CL_size_t, len(device))
	var stats = make([]CL_int, len(device))
	for i := 0; i < len(device); i++ {
		lens[i] = CL_size_t(len(binbts[i]))

	}
	prog := CLCreateProgramWithBinary(c.clContext, num, dids, lens, binbts, stats, &err)
	if err != CL_SUCCESS {
		return nil, fmt.Errorf("CLCreateProgramWithBinary Error code = %d", err)
	}
	return &Program{
		clProgram: prog,
		devices:   device,
	}, nil
}

func (c *Context) CreateProgramWithSource(sources []string) (*Program, error) {
	var num = len(sources)
	var err CL_int
	var count = CL_uint(num)
	var codes = make([][]byte, num)
	var lens = make([]CL_size_t, num)
	for i := 0; i < num; i++ {
		codes[i] = []byte(sources[i])
		lens[i] = CL_size_t(len(sources[i]))
	}
	prog := CLCreateProgramWithSource(c.clContext, count, codes, lens, &err)
	if err != CL_SUCCESS {
		return nil, fmt.Errorf("CLCreateProgramWithSource Error")
	}
	return &Program{
		clProgram: prog,
		devices:   c.devices,
	}, nil
}

func (c *Context) CreateCommandQueue(device *Device, properties int) (*CommandQueue, error) {
	var err CL_int
	queue := CLCreateCommandQueue(c.clContext, device.id, CL_command_queue_properties(properties), &err)
	if err != CL_SUCCESS {
		return nil, fmt.Errorf("CLCreateCommandQueue Error Code = %d", err)
	}
	return &CommandQueue{
		clQueue: queue,
		device:  device,
	}, nil
}

func (c *Context) CreateEmptyBuffer(flags CL_mem_flags, size int) (*MemObject, error) {
	var err CL_int
	mem := CLCreateBuffer(c.clContext, flags, CL_size_t(size), nil, &err)
	if err != CL_SUCCESS {
		return nil, fmt.Errorf("CLCreateBuffer Error Code = %d", err)
	}
	return &MemObject{
		clMem: mem,
	}, nil
}

func (p *Program) BuildProgram(dvs []*Device, opts string) error {
	var dnum = CL_uint(len(dvs))
	var dids = getDeviceIds(dvs)
	// []byte("-cl-kernel-arg-info")
	var optBuffer bytes.Buffer
	optBuffer.WriteString("-cl-std=CL1.2 -cl-kernel-arg-info ")
	optBuffer.WriteString(opts)
	ret := CLBuildProgram(p.clProgram, dnum, dids, optBuffer.Bytes(), nil, nil)
	if ret != CL_SUCCESS {
		if ret == CL_BUILD_PROGRAM_FAILURE {
			// print log
			var logsize CL_size_t
			CLGetProgramBuildInfo(p.clProgram, dids[0], CL_PROGRAM_BUILD_LOG, 0, nil, &logsize)
			var logcon interface{}
			logcon = make([]byte, logsize)
			CLGetProgramBuildInfo(p.clProgram, dids[0], CL_PROGRAM_BUILD_LOG, logsize, &logcon, &logsize)
			fmt.Println("CLBuildProgram Error: \n\n", logcon.(string))
		}
		return fmt.Errorf("CLBuildProgram Error Code = %d", ret)
	}
	return nil
}

func (p *Program) GetBinarieByDevices(devices []*Device) ([][]byte, error) {
	var dvdnum = len(devices)
	// get size
	var rtsz CL_size_t
	CLGetProgramInfo(p.clProgram, CL_PROGRAM_BINARY_SIZES, 0, nil, &rtsz)
	//fmt.Println("CLGetProgramInfo", rtsz)
	var val interface{}
	val = make([]CL_size_t, dvdnum)
	ret := CLGetProgramInfo(p.clProgram, CL_PROGRAM_BINARY_SIZES, rtsz, &val, &rtsz)
	if ret != CL_SUCCESS {
		return nil, fmt.Errorf("CLGetProgramInfo CL_PROGRAM_BINARY_SIZES Error")
	}
	//fmt.Println("CLGetProgramInfo CL_PROGRAM_BINARY_SIZES : ", val.([]CL_size_t))
	var binsizes = val.([]CL_size_t)
	var rtsize C.size_t
	var bins = make([]*C.char, dvdnum)
	fmt.Println("OpenCL Program GetBinarieByDevices dvdnum:", dvdnum, binsizes)
	for i := 0; i < dvdnum; i++ {
		bins[i] = (*C.char)(C.malloc(C.size_t(binsizes[i])))
	}
	var parmsize = dvdnum * int(unsafe.Sizeof(rtsz))
	fmt.Println("bins[]:", bins, "parmsize:", parmsize)
	c_errcode_ret := C.clGetProgramInfo(p.clProgram.cl_program,
		C.CL_PROGRAM_BINARIES,
		C.size_t(parmsize),
		unsafe.Pointer(&bins[0]),
		&rtsize)
	//fmt.Println("c_errcode_ret", c_errcode_ret, bins)
	if c_errcode_ret != CL_SUCCESS {
		return nil, fmt.Errorf("CLGetProgramInfo CL_PROGRAM_BINARIES Error code = %d", ret)
	}
	//fmt.Println("CLGetProgramInfo CL_PROGRAM_BINARIES : ", binsz, binary)
	var binbts = make([][]byte, dvdnum)
	for i := 0; i < dvdnum; i++ {
		binbts[i] = []byte(C.GoString(bins[i]))
	}
	//C.free(unsafe.Pointer(&bins[0]))
	//fmt.Println("binbts", binbts)
	//var binrets = *(*[]byte)(unsafe.Pointer(&bins[0]))
	// ok
	return binbts, nil
}

func (p *Program) CreateKernel(name string) (*Kernel, error) {
	var err CL_int
	kl := CLCreateKernel(p.clProgram, []byte(name), &err)
	if err != CL_SUCCESS {
		return nil, fmt.Errorf("CLCreateKernel Error Code = %d", err)
	}
	return &Kernel{
		clKernel: kl,
		name:     name,
	}, nil
}

func (p *Platform) GetDevices(ty CL_device_type) []*Device {
	var num CL_uint
	CLGetDeviceIDs(p.id, ty, 0, nil, &num)
	var dvcids = make([]CL_device_id, num)
	CLGetDeviceIDs(p.id, ty, num, dvcids, &num)
	devices := make([]*Device, num)
	for i := 0; i < int(num); i++ {
		var did = dvcids[i]
		var retsize CL_size_t
		CLGetDeviceInfo(did, CL_DEVICE_NAME, 0, nil, &retsize)
		var dname interface{}
		dname = make([]byte, retsize)
		CLGetDeviceInfo(did, CL_DEVICE_NAME, retsize, &dname, &retsize)
		devices[i] = &Device{id: did, name: dname.(string)}
	}
	return devices
}

func GetPlatforms() []*Platform {
	// start
	var pltmax CL_uint = 16
	var pltids = make([]CL_platform_id, pltmax)
	var pltnum CL_uint
	CLGetPlatformIDs(pltmax, pltids, &pltnum)
	platforms := make([]*Platform, pltnum)
	for i := 0; i < int(pltnum); i++ {
		var ptid = pltids[i]
		var retsize CL_size_t
		CLGetPlatformInfo(ptid, CL_PLATFORM_NAME, 0, nil, &retsize)
		var ptname interface{}
		ptname = make([]byte, retsize)
		//var ptnameptr interface{} = ptname
		CLGetPlatformInfo(ptid, CL_PLATFORM_NAME, retsize, (&ptname), &retsize)
		//fmt.Println("retsize: ", retsize)
		platforms[i] = &Platform{
			id:   ptid,
			name: ptname.(string),
		}
	}
	return platforms[0:pltnum]
}

func CreateContext(devices []*Device) (*Context, error) {
	var dids = getDeviceIds(devices)
	var err CL_int
	ctx := CLCreateContext(nil, CL_uint(len(devices)), dids, nil, nil, &err)
	if err != CL_SUCCESS {
		return nil, fmt.Errorf("CLCreateContext Error")
	}
	return &Context{
		clContext: ctx,
		devices:   devices,
	}, nil
}

func getDeviceIds(devices []*Device) []CL_device_id {
	dids := make([]CL_device_id, len(devices))
	for i := 0; i < len(devices); i++ {
		dids[i] = devices[i].id
	}
	return dids
}
