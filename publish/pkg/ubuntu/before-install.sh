#!/bin/sh

# create dhound-agent group
if ! getent group dhound-agent >/dev/null; then
  groupadd -r dhound-agent
fi

# create dhound-agent user
if ! getent passwd dhound-agent >/dev/null; then
  useradd -M -r -g dhound-agent -d /var/lib/dhound-agent \
    -s /usr/sbin/nologin -c "dhound-agent Service User" dhound-agent
fi
