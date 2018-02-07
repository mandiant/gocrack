#!/bin/bash

echo $PWD

cd /home/travis/
git clone https://github.com/hashcat/hashcat.git && cd hashcat && git submodule init && git submodule update && git checkout v3.6.0
make SHARED=1