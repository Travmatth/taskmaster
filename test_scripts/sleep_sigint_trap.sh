#!/bin/bash

trap -- "echo -n INT caught; trap INT" INT
while :
do
sleep 1
done
exit 0