@echo off
rem run this script as admin

sc create dhound-agent binpath="\"d:\_DEV\GIT\dhound-agent\dhound-agent.exe\" -log-file \"d:\_DEV\GIT\dhound-agent\log\dhound-agent.log\" -config-dir \"d:\_DEV\GIT\dhound-agent\config\" -verbose" start=auto DisplayName="Dhound.io agent"
sc description dhound-agent "Collects security events in the system for further analytics in dhound.io service"
sc start dhound-agent
sc query dhound-agent

echo Check dhound-agent log files
:exit