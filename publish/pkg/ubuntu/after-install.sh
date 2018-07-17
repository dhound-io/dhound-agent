#!/bin/sh

chown -R dhound-agent:dhound-agent /opt/dhound-agent
chown dhound-agent /var/log/dhound-agent
chown dhound-agent:dhound-agent /var/lib/dhound-agent

update-rc.d dhound-agent defaults

echo "Logs for dhound-agent will be in /var/log/dhound-agent/"
