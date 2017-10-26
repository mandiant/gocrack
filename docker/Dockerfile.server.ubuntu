FROM ubuntu:xenial

ARG USER_ID
ARG AUTHOR

RUN apt-get update && apt-get install -y --no-install-recommends \
        ocl-icd-opencl-dev && \
    rm -rf /var/lib/apt/lists/*

RUN echo "/usr/local/cuda/lib64" >> /etc/ld.so.conf.d/cuda.conf && \
    ldconfig

# nvidia-docker 1.0
LABEL com.nvidia.volumes.needed="nvidia_driver"

RUN echo "/usr/local/nvidia/lib" >> /etc/ld.so.conf.d/nvidia.conf && \
    echo "/usr/local/nvidia/lib64" >> /etc/ld.so.conf.d/nvidia.conf

ENV PATH /usr/local/nvidia/bin:/usr/local/cuda/bin:${PATH}
ENV LD_LIBRARY_PATH /usr/local/nvidia/lib:/usr/local/nvidia/lib64

COPY dist/gocrack/gocrack_server /usr/local/bin/gocrack_server
COPY dist/hashcat/bin/hashcat /usr/local/bin
COPY dist/hashcat/lib/libhashcat.so /usr/local/lib
COPY dist/hashcat/share /usr/local/share
COPY files/server_entrypoint.sh /usr/local/bin/entrypoint.sh

RUN ldconfig

ENTRYPOINT [ "bash", "/usr/local/bin/entrypoint.sh" ]