#!/bin/sh

if [ ! -z $(type -P ssh) ]; then
	echo "true"
else
	echo "false"
fi