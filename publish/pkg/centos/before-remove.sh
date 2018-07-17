if [ $1 -eq 0 ]; then
  /sbin/service dhound-agent stop >/dev/null 2>&1 || true
  /sbin/chkconfig --del dhound-agent
  if getent passwd dhound-agent >/dev/null ; then
    userdel dhound-agent
  fi

  if getent group dhound-agent > /dev/null ; then
    groupdel dhound-agent
  fi
fi
