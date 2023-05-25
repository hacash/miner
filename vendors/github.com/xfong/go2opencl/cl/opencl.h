/*
  This file is used to point the compiler to the actual opencl.h of the system.
  It is also used to check the version of opencl installed
*/
#define CL_TARGET_OPENCL_VERSION 300
#include <stdlib.h>
#define CL_USE_DEPRECATED_OPENCL_1_2_APIS
#define CL_USE_DEPRECATED_OPENCL_2_0_APIS
#ifdef __APPLE__
	#include <OpenCL/OpenCL.h>
#else
	#include <CL/opencl.h>
#ifdef __WIN32
	#include <CL/cl_dx9_media_sharing.h>
	#include <CL/cl_d3d10.h>
	#include <CL/cl_d3d11.h>
#endif
#endif

#ifndef CL_VERSION_1_2
	#error "This package requires OpenCL 1.2"
#endif

