#!/bin/sh

if [ ! -z $(type -P dnsmasq) ]; then
	echo "true"
else
	echo "false"
fi