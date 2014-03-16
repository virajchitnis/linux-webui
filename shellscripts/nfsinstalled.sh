#!/bin/sh

if [ ! -z $(type -P exportfs) ]; then
	echo "true"
else
	echo "false"
fi