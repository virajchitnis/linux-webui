#!/bin/sh

if [ ! -z $(type -P smbpasswd) ]; then
	echo "true"
else
	echo "false"
fi