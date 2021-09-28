#!/bin/bash

if [ $1 -eq 1 ] ; then 
  # Enable the service
  systemctl link /opt/go-server/etc/go-server.service
  systemctl enable go-server.service
else 
  # Reload services and wait for startup
  systemctl daemon-reload
fi

# Restart server
systemctl restart go-server.service

