#!/bin/sh

if [ -d /etc/cups ]; then
	echo "true"
else
	echo "false"
fi