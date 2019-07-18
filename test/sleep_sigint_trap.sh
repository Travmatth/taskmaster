#!/bin/bash

echo here
while :
do
trap -- "echo INT caught; trap INT" INT
sleep 1
done
exit 0