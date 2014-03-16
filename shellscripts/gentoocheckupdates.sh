#!/bin/sh

eselect news list
emerge -pv --update --deep --with-bdeps=y --newuse @world