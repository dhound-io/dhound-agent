#!/bin/sh

REPOFOLDER=/var/dhound.io/repository

# publish debian packages
find -name \*.deb -exec reprepro --ignore=undefinedtarget -Vb $REPOFOLDER/deb includedeb dhound-agent {} \;

# public rpm packages
cp *.rpm $REPOFOLDER/rpm/
createrepo --outputdir=$REPOFOLDER/rpm/ . --update

sudo chmod -R ugo+rX $REPOFOLDER

