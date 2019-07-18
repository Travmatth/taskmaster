#!/bin/bash
FILE=test/taskmaster_starttimeout_tmp
if [ ! -f "$FILE" ]; then
    touch $FILE
    exit 1
fi
sleep 3
rm $FILE
exit 0;