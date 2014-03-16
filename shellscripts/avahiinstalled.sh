#!/bin/sh

if [ -d /etc/avahi ]; then
	echo "true"
else
	echo "false"
fi