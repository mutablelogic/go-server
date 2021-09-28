#!/bin/bash

# Add nginx user
id -u nginx &>/dev/null || useradd --system nginx

# Change permissions on var folder
if [ -d "/opt/go-server/var" ] ; then
  chown -R nginx "/opt/go-server/var"
fi

# Enable service
systemctl link /opt/go-server/etc/go-server.service
systemctl enable go-server.service

# Restart server
systemctl restart go-server.service
