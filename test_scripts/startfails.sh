#!/bin/bash
FILE=test_scripts/taskmaster_startfail_tmp
if [ ! -f "$FILE" ]; then
    touch $FILE
    echo '1' > $FILE
    exit 1
fi
exec 5<>$FILE
read -u 5 -n 1 VAL
exec 5>&-
echo $VAL
if [ $VAL = "5" ]; then
    rm $FILE
    exit 0;
else
    echo "$(($VAL + 1))" > $FILE
    exit 1;
fi