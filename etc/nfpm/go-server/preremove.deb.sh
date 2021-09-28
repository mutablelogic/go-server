#!/bin/bash
if [ $1 -eq 0 ] ; then 
  systemctl --no-reload disable go-server.service
  systemctl stop go-server.service
fi
