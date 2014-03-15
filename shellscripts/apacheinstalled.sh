#!/bin/sh

if [ -d /etc/apache2 ] || [ -d /etc/httpd ]; then
	echo "true"
else
	echo "false"
fi