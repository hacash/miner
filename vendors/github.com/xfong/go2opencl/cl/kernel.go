package cl

/*
#include "./opencl.h"

extern void go_native_kernel(void *user_args);
static void CL_CALLBACK c_enqueue_native_kernel(void *user_args) {
        go_native_kernel((void *)user_args);
}

static cl_int CLEnqueueNativeKernel(      cl_command_queue command_queue,
                                                        void *                                  user_args,
							size_t					num_args,
							cl_uint					num_mem_objects,
						const cl_mem *				mem_list,
						const void **				args_mem_ptrs,
							cl_uint					num_events_in_list,
						const cl_event *				eventsWaitList,
                                                        cl_event *                                ret_event){
        return clEnqueueNativeKernel(command_queue, c_enqueue_native_kernel, user_args, num_args, num_mem_objects, mem_list, args_mem_ptrs, num_events_in_list, eventsWaitList, ret_event);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

//////////////// Basic Types ////////////////
type ErrUnsupportedArgumentType struct {
	Index int
	Value interface{}
}

func (e ErrUnsupportedArgumentType) Error() string {
	return fmt.Sprintf("cl: unsupported argument type for index %d: %+v", e.Index, e.Value)
}

//////////////// Abstract Types ////////////////
type Kernel struct {
	clKernel C.cl_kernel
	name     string
}

//////////////// Golang Types ////////////////
type LocalBuffer int

////////////////// Supporting Types ////////////////
type CL_go_native_kernel func(user_data unsafe.Pointer)

var go_native_kernel_func map[unsafe.Pointer]CL_go_native_kernel

//////////////// Basic Functions ////////////////
func init() {
	go_native_kernel_func = make(map[unsafe.Pointer]CL_go_native_kernel)
}

//export go_native_kernel
func go_native_kernel(user_data unsafe.Pointer) {
	var c_user_data []unsafe.Pointer
	c_user_data = *(*[]unsafe.Pointer)(user_data)
	go_native_kernel_func[c_user_data[1]](c_user_data[0])
}

func releaseKernel(k *Kernel) {
	if k.clKernel != nil {
		C.clReleaseKernel(k.clKernel)
		k.clKernel = nil
	}
}

func retainKernel(k *Kernel) {
	if k.clKernel != nil {
		C.clRetainKernel(k.clKernel)
	}
}

//////////////// Abstract Functions ////////////////
func (k *Kernel) Release() {
	releaseKernel(k)
}

func (k *Kernel) Retain() {
	retainKernel(k)
}

func (k *Kernel) SetArgs(args ...interface{}) error {
	for index, arg := range args {
		if err := k.SetArg(index, arg); err != nil {
			return err
		}
	}
	return nil
}

func (k *Kernel) SetArg(index int, arg interface{}) error {
	switch val := arg.(type) {
	case uint8:
		return k.SetArgUint8(index, val)
	case int8:
		return k.SetArgInt8(index, val)
	case uint32:
		return k.SetArgUint32(index, val)
	case uint64:
		return k.SetArgUint64(index, val)
	case int32:
		return k.SetArgInt32(index, val)
	case float32:
		return k.SetArgFloat32(index, val)
	case *MemObject:
		return k.SetArgBuffer(index, val)
	case LocalBuffer:
		return k.SetArgLocal(index, int(val))
	default:
		return ErrUnsupportedArgumentType{Index: index, Value: arg}
	}
}

func (k *Kernel) ArgAddressQualifier(index int) (string, error) {
	var val C.cl_kernel_arg_address_qualifier
	var err C.cl_int
	defer C.free(unsafe.Pointer(&err))
	if err = C.clGetKernelArgInfo(k.clKernel, C.cl_uint(index), C.CL_KERNEL_ARG_ADDRESS_QUALIFIER, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return "", toError(err)
	}
	switch val {
	default:
		return "", toError(err)
	case C.CL_KERNEL_ARG_ADDRESS_GLOBAL:
		return "Global", nil
	case C.CL_KERNEL_ARG_ADDRESS_LOCAL:
		return "Local", nil
	case C.CL_KERNEL_ARG_ADDRESS_CONSTANT:
		return "Constant", nil
	case C.CL_KERNEL_ARG_ADDRESS_PRIVATE:
		return "Private", nil
	}
}

func (k *Kernel) ArgAccessQualifier(index int) (string, error) {
	var val C.cl_kernel_arg_access_qualifier
	var err C.cl_int
	defer C.free(unsafe.Pointer(&err))
	if err = C.clGetKernelArgInfo(k.clKernel, C.cl_uint(index), C.CL_KERNEL_ARG_ACCESS_QUALIFIER, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return "", toError(err)
	}
	switch val {
	default:
		return "", toError(err)
	case C.CL_KERNEL_ARG_ACCESS_READ_ONLY:
		return "ReadOnly", nil
	case C.CL_KERNEL_ARG_ACCESS_READ_WRITE:
		return "ReadWrite", nil
	case C.CL_KERNEL_ARG_ACCESS_WRITE_ONLY:
		return "WriteOnly", nil
	case C.CL_KERNEL_ARG_ACCESS_NONE:
		return "None", nil
	}
}

func (k *Kernel) ArgTypeQualifier(index int) (C.cl_kernel_arg_type_qualifier, error) {
	var val C.cl_kernel_arg_type_qualifier
	err := C.clGetKernelArgInfo(k.clKernel, C.cl_uint(index), C.CL_KERNEL_ARG_TYPE_QUALIFIER, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil)
	return val, toError(err)
}

func (k *Kernel) ArgName(index int) (string, error) {
	var strC [1024]byte
	var strN C.size_t
	if err := C.clGetKernelArgInfo(k.clKernel, C.cl_uint(index), C.CL_KERNEL_ARG_NAME, 1024, unsafe.Pointer(&strC[0]), &strN); err != C.CL_SUCCESS {
		return "", toError(err)
	}
	return string(strC[:strN]), nil
}

func (k *Kernel) ArgTypeName(index int) (string, error) {
	var strC [1024]byte
	var strN C.size_t
	if err := C.clGetKernelArgInfo(k.clKernel, C.cl_uint(index), C.CL_KERNEL_ARG_TYPE_NAME, 1024, unsafe.Pointer(&strC[0]), &strN); err != C.CL_SUCCESS {
		return "", toError(err)
	}
	return string(strC[:strN]), nil
}

func (k *Kernel) SetArgBuffer(index int, buffer *MemObject) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(buffer.clMem)), unsafe.Pointer(&buffer.clMem))
}

func (k *Kernel) SetArgFloat32(index int, val float32) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

func (k *Kernel) SetArgInt8(index int, val int8) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

func (k *Kernel) SetArgUint8(index int, val uint8) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

func (k *Kernel) SetArgInt32(index int, val int32) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

func (k *Kernel) SetArgUint32(index int, val uint32) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

func (k *Kernel) SetArgUint64(index int, val uint64) error {
	return k.SetArgUnsafe(index, int(unsafe.Sizeof(val)), unsafe.Pointer(&val))
}

func (k *Kernel) SetArgLocal(index int, size int) error {
	return k.SetArgUnsafe(index, size, nil)
}

func (k *Kernel) SetArgUnsafe(index, argSize int, arg unsafe.Pointer) error {
	//fmt.Println("FUNKY: ", index, argSize)
	return toError(C.clSetKernelArg(k.clKernel, C.cl_uint(index), C.size_t(argSize), arg))
}

func (k *Kernel) GlobalWorkGroupSize(device *Device) ([3]int, error) {
	var size [3]C.size_t
	if err := C.clGetKernelWorkGroupInfo(k.clKernel, device.nullableId(), C.CL_KERNEL_GLOBAL_WORK_SIZE, C.size_t(unsafe.Sizeof(size)), unsafe.Pointer(&size[0]), nil); err != C.CL_SUCCESS {
		return [3]int{-1, -1, -1}, toError(err)
	}
	return [3]int{int(size[0]), int(size[1]), int(size[2])}, nil
}

func (k *Kernel) WorkGroupSize(device *Device) (int, error) {
	var size C.size_t
	err := C.clGetKernelWorkGroupInfo(k.clKernel, device.nullableId(), C.CL_KERNEL_WORK_GROUP_SIZE, C.size_t(unsafe.Sizeof(size)), unsafe.Pointer(&size), nil)
	return int(size), toError(err)
}

func (k *Kernel) PreferredWorkGroupSizeMultiple(device *Device) (int, error) {
	var size C.size_t
	err := C.clGetKernelWorkGroupInfo(k.clKernel, device.nullableId(), C.CL_KERNEL_PREFERRED_WORK_GROUP_SIZE_MULTIPLE, C.size_t(unsafe.Sizeof(size)), unsafe.Pointer(&size), nil)
	return int(size), toError(err)
}

func (k *Kernel) CompileWorkGroupSize(device *Device) ([3]int, error) {
	var wgSize [3]C.size_t
	defer C.free(unsafe.Pointer(&wgSize))
	if err := C.clGetKernelWorkGroupInfo(k.clKernel, device.nullableId(), C.CL_KERNEL_COMPILE_WORK_GROUP_SIZE, C.size_t(unsafe.Sizeof(wgSize)), unsafe.Pointer(&wgSize), nil); err != C.CL_SUCCESS {
		return [3]int{-1, -1, -1}, toError(err)
	}
	return [3]int{int(wgSize[0]), int(wgSize[1]), int(wgSize[2])}, nil
}

func (k *Kernel) WorkGroupLocalMemSize(device *Device) (int, error) {
	var size C.size_t
	err := C.clGetKernelWorkGroupInfo(k.clKernel, device.nullableId(), C.CL_KERNEL_LOCAL_MEM_SIZE, C.size_t(unsafe.Sizeof(size)), unsafe.Pointer(&size), nil)
	return int(size), toError(err)
}

func (k *Kernel) WorkGroupPrivateMemSize(device *Device) (int, error) {
	var size C.size_t
	err := C.clGetKernelWorkGroupInfo(k.clKernel, device.nullableId(), C.CL_KERNEL_PRIVATE_MEM_SIZE, C.size_t(unsafe.Sizeof(size)), unsafe.Pointer(&size), nil)
	return int(size), toError(err)
}

func (k *Kernel) NumArgs() (int, error) {
	var num C.cl_uint
	err := C.clGetKernelInfo(k.clKernel, C.CL_KERNEL_NUM_ARGS, C.size_t(unsafe.Sizeof(num)), unsafe.Pointer(&num), nil)
	return int(num), toError(err)
}

func (k *Kernel) ReferenceCount() (int, error) {
	var num C.cl_uint
	err := C.clGetKernelInfo(k.clKernel, C.CL_KERNEL_REFERENCE_COUNT, C.size_t(unsafe.Sizeof(num)), unsafe.Pointer(&num), nil)
	return int(num), toError(err)
}

func (k *Kernel) FunctionName() (string, error) {
	var name C.char
	err := C.clGetKernelInfo(k.clKernel, C.CL_KERNEL_FUNCTION_NAME, C.size_t(unsafe.Sizeof(name)), unsafe.Pointer(&name), nil)
	return C.GoString(&name), toError(err)
}

func (k *Kernel) Attributes() (string, error) {
	var name C.char
	err := C.clGetKernelInfo(k.clKernel, C.CL_KERNEL_ATTRIBUTES, C.size_t(unsafe.Sizeof(name)), unsafe.Pointer(&name), nil)
	return C.GoString(&name), toError(err)
}

func (k *Kernel) Context() (*Context, error) {
	var context C.cl_context
	err := C.clGetKernelInfo(k.clKernel, C.CL_KERNEL_CONTEXT, C.size_t(unsafe.Sizeof(context)), unsafe.Pointer(&context), nil)
	return &Context{clContext: context, devices: nil}, toError(err)
}

func (k *Kernel) Program() (*Program, error) {
	var program C.cl_program
	err := C.clGetKernelInfo(k.clKernel, C.CL_KERNEL_PROGRAM, C.size_t(unsafe.Sizeof(program)), unsafe.Pointer(&program), nil)
	return &Program{clProgram: program, devices: nil}, toError(err)
}

// Enqueues a command to execute a kernel on a device.
func (q *CommandQueue) EnqueueNDRangeKernel(kernel *Kernel, globalWorkOffset, globalWorkSize, localWorkSize []int, eventWaitList []*Event) (*Event, error) {
	workDim := len(globalWorkSize)
	var globalWorkOffsetList []C.size_t
	var globalWorkOffsetPtr *C.size_t
	if globalWorkOffset != nil {
		globalWorkOffsetList = make([]C.size_t, len(globalWorkOffset))
		for i, off := range globalWorkOffset {
			globalWorkOffsetList[i] = C.size_t(off)
		}
		globalWorkOffsetPtr = &globalWorkOffsetList[0]
	}
	var globalWorkSizeList []C.size_t
	var globalWorkSizePtr *C.size_t
	if globalWorkSize != nil {
		globalWorkSizeList = make([]C.size_t, len(globalWorkSize))
		for i, off := range globalWorkSize {
			globalWorkSizeList[i] = C.size_t(off)
		}
		globalWorkSizePtr = &globalWorkSizeList[0]
	}
	var localWorkSizeList []C.size_t
	var localWorkSizePtr *C.size_t
	if localWorkSize != nil {
		localWorkSizeList = make([]C.size_t, len(localWorkSize))
		for i, off := range localWorkSize {
			localWorkSizeList[i] = C.size_t(off)
		}
		localWorkSizePtr = &localWorkSizeList[0]
	}
	var event C.cl_event
	err := toError(C.clEnqueueNDRangeKernel(q.clQueue, kernel.clKernel, C.cl_uint(workDim), globalWorkOffsetPtr, globalWorkSizePtr, localWorkSizePtr, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

// Enqueues a command to execute a kernel on a device, except with globalWorkSize = localWorkSize = 1
// and globalWorkOffset = 0
func (q *CommandQueue) EnqueueTask(kernel *Kernel, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	err := toError(C.clEnqueueTask(q.clQueue, kernel.clKernel, C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

// Enqueues a native user function for execution on on a device. Need CL_EXEC_NATIVE_KERNEL capability to be present.
func (q *CommandQueue) EnqueueNativeKernel(user_args unsafe.Pointer, num_user_args int, memObjects []*MemObject, ptr_memobj_in_args []unsafe.Pointer, eventWaitList []*Event) (*Event, error) {
	var event C.cl_event
	UserMemObjs := make([]C.cl_mem, len(memObjects))
	for i, mb := range memObjects {
		UserMemObjs[i] = mb.clMem
	}
	err := toError(C.CLEnqueueNativeKernel(q.clQueue, user_args, C.size_t(num_user_args), C.cl_uint(len(memObjects)), &UserMemObjs[0], &ptr_memobj_in_args[0], C.cl_uint(len(eventWaitList)), eventListPtr(eventWaitList), &event))
	return newEvent(event), err
}

func (p *Program) CreateKernelsInProgram() ([]*Kernel, error) {
	var num_kerns C.cl_uint
	err := C.clCreateKernelsInProgram(p.clProgram, 1, nil, &num_kerns)
	if toError(err) != nil {
		fmt.Printf("Error getting number of kernels to create \n")
		return nil, toError(err)
	}
	kernel_list := make([]C.cl_kernel, int(num_kerns))
	err = C.clCreateKernelsInProgram(p.clProgram, num_kerns, &kernel_list[0], nil)
	if toError(err) != nil {
		fmt.Printf("Error creating kernels \n")
		return nil, toError(err)
	}
	returnKerns := make([]*Kernel, len(kernel_list))
	for i, kptr := range kernel_list {
		testKern := &Kernel{clKernel: kptr, name: ""}
		kname, errK := testKern.FunctionName()
		if errK == nil {
			returnKerns[i].clKernel = kptr
			returnKerns[i].name = kname
		} else {
			fmt.Printf("Error getting information about kernel %d \n", i)
			returnKerns[i] = nil
		}
	}
	return returnKerns, nil
}
