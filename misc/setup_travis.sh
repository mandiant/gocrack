#!/bin/bash


cd $HOME
git clone https://github.com/hashcat/hashcat.git && cd $HOME/hashcat && git submodule update --init --recursive && git checkout $HASHCAT_VERSION
make SHARED=1
