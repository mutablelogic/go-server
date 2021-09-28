#!/bin/bash

# Daemon reload
systemctl --system daemon-reload

# Stop service
deb-systemd-invoke stop go-server.service

# Purge
deb-systemd-helper purge go-server.service
deb-systemd-helper unmask go-server.service

