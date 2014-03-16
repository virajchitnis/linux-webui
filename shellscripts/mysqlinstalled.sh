#!/bin/sh

if [ -d /etc/mysql ] || [ ! -z $(type -P mysql) ]; then
	echo "true"
else
	echo "false"
fi