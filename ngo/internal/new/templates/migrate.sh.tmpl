#!/bin/sh
set -u

before=$(goose version 2>&1 | awk '{print $NF}')
case "$before" in
'' | *[!0-9]*)
	echo "cannot read current version: $before"
	exit 1
	;;
esac

goose up && exit 0

# the failing file already rolled itself back; this undoes the files applied before it in this run
echo "migration failed, rolling back to $before"
goose down-to "$before"
exit 1
