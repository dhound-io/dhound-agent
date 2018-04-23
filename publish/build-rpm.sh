#!/bin/sh

cd ..

env GOOS=linux GOARCH=arm GOARM=5 ARCHITECTURE=armel make rpm
env GOOS=linux GOARCH=arm ARCHITECTURE=arm make rpm
env GOOS=linux GOARCH=arm ARCHITECTURE=armhf make rpm
env GOOS=linux GOARCH=arm GOARM=6 ARCHITECTURE=armv6hl make rpm
env GOOS=linux GOARCH=arm GOARM=7 ARCHITECTURE=armv7hl make rpm
env GOOS=linux GOARCH=arm64 ARCHITECTURE=arm64 make rpm

env CC=gcc GOOS=linux GOARCH=386 CGO_ENABLED=1 GOGCCFLAGS="-fPIC -m32 -fmessage-length=0" ARCHITECTURE=i386 CGO_LDFLAGS="-L/dhound-build-server/dhound-agent/publish/libpcap/i386"  make rpm
env CC=gcc GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOGCCFLAGS="-fPIC -m64 -fmessage-length=0" ARCHITECTURE=amd64 make rpm
env CC=gcc GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOGCCFLAGS="-fPIC -m64 -fmessage-length=0" ARCHITECTURE=ia64 make rpm

#env GOOS=linux GOARCH=386 ARCHITECTURE=i386 make rpm
#env GOOS=linux GOARCH=amd64 ARCHITECTURE=amd64 make rpm
#env GOOS=linux GOARCH=amd64 ARCHITECTURE=ia64 make rpm

mv *.rpm publish/

cd publish/

echo 'Sign rpm packages'
for line in $(find . -iname '*.rpm'); do
    ./sign-rpm.sh ""  "$line"
   rpm --checksig "$line"
done



