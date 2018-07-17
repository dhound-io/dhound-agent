#!/bin/sh

./clean.sh
./build-debian.sh
./build-rpm.sh
./publish-into-repo.sh
