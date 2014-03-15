#!/bin/sh

if [ $(uname -s) == 'Linux' ]; then
	if [ ! -z $(type -P emerge) ]; then
		echo "Gentoo"
	fi
	
	if [ ! -z $(type -P apt-get) ]; then
		echo "Debian"
	fi
	
	if [ ! -z $(type -P yum) ]; then
		echo "CentOS"
	fi
fi