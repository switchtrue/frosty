#!/bin/sh

echoerr() { echo "$@" 1>&2; }

echo "Frosty Job Dir: $FROSTY_JOB_DIR"
echo "Frosty Artifacts Dir: $FROSTY_JOB_ARTIFACTS_DIR"

>&2 echo "Uh-Oh! Its all gone wrong!"

echoerr hello world

exit 1