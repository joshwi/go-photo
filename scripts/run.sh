#!/bin/bash

while getopts "s:t:f:" arg; do
  case $arg in
    s) source=$OPTARG;;
    t) target=$OPTARG;;
    f) filetypes=$OPTARG;;
  esac
done

# cd /go-media

(
set -e
app/builds/transfer -source "$source" -target "$target" -files "$filetypes"
app/builds/read
app/builds/transactions
app/builds/audit
)

if [ "$?" -ne 0 ]; then
  echo "The world is on fire!"
fi