#!/bin/bash

echo $PWD

cd /home/travis/
git clone https://github.com/hashcat/hashcat.git && cd hashcat && git submodule update --init --recursive && git checkout v3.6.0
make SHARED=1
make install

cp -r include/ /usr/local/include/hashcat
cp -r deps/OpenCL-Headers/CL /usr/local/include/CL
