# Building GoCrack

**Note**: Use docker? We have [images for containers](../docker/README.md) for both building and running gocrack.

## Prerequisites

* Linux (Ubuntu 16.04+ although other distributions may work) or MacOS
* Computer(s) with NVIDIA or AMD GPUs

## Step 1. Building Hashcat

1. GoCrack requires hashcat version 3.6 or higher and to be built in the `Shared` mode. This can be accomplished by switching the `SHARED` bit to 1 in `src/Makefile`. Alternatively, you can apply the patch [here](../docker/files/hashcat_shared.patch).
1. Follow Hashcat's [build instructions](https://github.com/hashcat/hashcat/blob/master/BUILD.md) to compile hashcat as a shared library. 
1. Copy hashcat's `include` folder to `/usr/local/include/hashcat` -> `cp -r include/ /usr/local/include/hashcat`
1. Test and ensure hashcat was installed successfully by running `hashcat --opencl-info`. You should see information about the various OpenCL devices attached to your computer.

## Step 2. Building GoCrack

### MacOS

1. Building GoCrack's server & workers is straightforward on MacOS as libOpenCL is a "standard" framework. Simply run, `make build`. 

### Linux

You'll most likely need to install libOpenCL along with platform specific ICD's for all your devices to work. At the time of writing, we only have access to NVIDIA GPUs.

To test and ensure OpenCL libraries are working correctly you can run `clinfo` to show all OpenCL platforms and devices on your machine.

```
$ clinfo
Number of platforms                               2
  Platform Name                                   Intel(R) OpenCL
  Platform Vendor                                 Intel(R) Corporation
  Platform Version                                OpenCL 1.2 LINUX
  Platform Profile                                FULL_PROFILE
  Platform Extensions                             cl_khr_icd cl_khr_global_int32_base_atomics cl_khr_global_int32_extended_atomics cl_khr_local_int32_base_atomics cl_khr_local_int32_extended_atomics cl_khr_byte_addressable_store cl_khr_depth_images cl_khr_3d_image_writes cl_intel_exec_by_local_thread cl_khr_spir cl_khr_fp64
  Platform Extensions function suffix             INTEL

  Platform Name                                   NVIDIA CUDA
  Platform Vendor                                 NVIDIA Corporation
  Platform Version                                OpenCL 1.2 CUDA 8.0.0
  Platform Profile                                FULL_PROFILE
  Platform Extensions                             cl_khr_global_int32_base_atomics cl_khr_global_int32_extended_atomics cl_khr_local_int32_base_atomics cl_khr_local_int32_extended_atomics cl_khr_fp64 cl_khr_byte_addressable_store cl_khr_icd cl_khr_gl_sharing cl_nv_compiler_options cl_nv_device_attribute_query cl_nv_pragma_unroll cl_nv_copy_opts
  Platform Extensions function suffix             NV
```

Building GoCrack's server & workers can now be accomplished by running `make build`.

## Build Tags

GoCrack is built to use whatever authentication backend and storage provider you choose. By default, all supported modules are compiled into binary but you have the option to exclude ones you do not want.

Example use:

    $ make SERVBUILDTAGS="auth_database auth_ldap"

### Authentication

1. `auth_database`: Allows you to use whatever storage backend you've chosen for authentication
1. `auth_ldap`: Allows you to use the LDAP authentication provider

### Database

1. `stor_bdb`: Build GoCrack with the BoltDB flatfile engine
