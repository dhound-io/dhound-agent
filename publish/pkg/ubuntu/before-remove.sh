#!/bin/sh

if [ $1 = "remove" ]; then
  service dhound-agent stop >/dev/null 2>&1 || true
  update-rc.d -f dhound-agent remove

  if getent passwd dhound-agent >/dev/null ; then
    userdel dhound-agent
  fi

  if getent group dhound-agent >/dev/null ; then
    groupdel dhound-agent
  fi
fi
