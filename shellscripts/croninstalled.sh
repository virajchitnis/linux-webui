#!/bin/sh

if [ ! -z $(type -P crontab) ]; then
	echo "true"
else
	echo "false"
fi