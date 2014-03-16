#!/bin/sh

if [ -d /etc/netatalk ] || [ -f /etc/afp.conf ]; then
	echo "true"
else
	echo "false"
fi