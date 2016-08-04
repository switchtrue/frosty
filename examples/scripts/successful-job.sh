#!/bin/sh

echo "Starting backup"

echo "Here is some text to backup" >> $FROSTY_JOB_ARTIFACTS_DIR/backup.txt
echo "One" >> $FROSTY_JOB_ARTIFACTS_DIR/file1.txt
echo "Two" >> $FROSTY_JOB_ARTIFACTS_DIR/file2.txt

mkdir $FROSTY_JOB_ARTIFACTS_DIR/nested
echo "Three" >> $FROSTY_JOB_ARTIFACTS_DIR/nested/file3.txt

echo "Backup complete"