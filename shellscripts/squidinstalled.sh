#!/bin/sh

if [ -d /etc/squid ] || [ -d /etc/squid3 ]; then
	echo "true"
else
	echo "false"
fi