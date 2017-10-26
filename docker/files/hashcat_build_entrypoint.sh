#/bin/bash

adduser --disabled-password --gecos "" --uid $USER_ID build
su build -c "make DESTDIR=/out PREFIX= install"
