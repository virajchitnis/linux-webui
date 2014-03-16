#!/bin/sh

if [ -f /etc/init.d/git-daemon ] || [ -f /etc/conf.d/git-daemon ]; then
	echo "true"
else
	echo "false"
fi