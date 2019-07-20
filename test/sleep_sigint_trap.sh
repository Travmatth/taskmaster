#!/bin/bash

trap -- "echo INT caught; trap INT" INT
while :
do
sleep 1
done
exit 0