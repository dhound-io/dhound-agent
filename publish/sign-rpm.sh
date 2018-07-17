#!/usr/bin/expect -f
# wrapper to make rpm --sign be non-interactive
# passwd is 1st arg, file to sign is 2nd
#send_user «$argv0 [lrange $argv 0 2]\n" 
#set files [lrange $argv 1 $argc ]

set password [lindex $argv 0]
set files [lrange $argv 1 end ]
#echo "$files"
spawn rpm --addsign $files
expect "Enter pass phrase:"
send -- "$password\r"
expect eof
