FROM nvidia/opencl:runtime-ubuntu18.04

RUN apt-get update && apt-get install -y --no-install-recommends \
        ocl-icd-opencl-dev && \
    rm -rf /var/lib/apt/lists/*

COPY dist/gocrack/gocrack_worker /usr/local/bin/gocrack_worker
COPY dist/hashcat/bin/hashcat /usr/local/bin
COPY dist/hashcat/lib/libhashcat.so /usr/local/lib
COPY dist/hashcat/share /usr/local/share
COPY files/worker_entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT [ "bash", "/usr/local/bin/entrypoint.sh" ]
