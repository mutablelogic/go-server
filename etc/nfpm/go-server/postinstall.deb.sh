#!/bin/bash

# Add nginx user
id -u nginx &>/dev/null || useradd --system nginx

# Enable service
systemctl link /opt/go-server/etc/go-server.service
systemctl enable go-server.service

# Restart server
systemctl restart go-server.service
