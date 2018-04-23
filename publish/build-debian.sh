#!/bin/sh

# check official debian ports https://www.debian.org/ports/

cd ..
#https://github.com/golang/go/wiki/GoArm

env GOOS=linux GOARCH=arm GOARM=5 ARCHITECTURE=armel make deb
env GOOS=linux GOARCH=arm ARCHITECTURE=arm make deb
env GOOS=linux GOARCH=arm ARCHITECTURE=armhf make deb
env GOOS=linux GOARCH=arm GOARM=6 ARCHITECTURE=armv6hl make deb
env GOOS=linux GOARCH=arm GOARM=7 ARCHITECTURE=armv7hl make deb
env GOOS=linux GOARCH=arm64 ARCHITECTURE=arm64 make deb


env CC=gcc GOOS=linux GOARCH=386 CGO_ENABLED=1 GOGCCFLAGS="-fPIC -m32 -fmessage-length=0" CGO_LDFLAGS="-L/dhound-build-server/dhound-agent/publish/libpcap/i386"  ARCHITECTURE=i386 make deb
env CC=gcc GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOGCCFLAGS="-fPIC -m64 -fmessage-length=0" ARCHITECTURE=amd64 make deb
env CC=gcc GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOGCCFLAGS="-fPIC -m64 -fmessage-length=0" ARCHITECTURE=ia64 make deb

mv *.deb publish/
