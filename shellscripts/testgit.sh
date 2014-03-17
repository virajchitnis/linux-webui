#!/bin/sh

# Check if the installed version of the application is a git repo

if [ -d .git ]; then
	echo "true"
else
	echo "false"
fi