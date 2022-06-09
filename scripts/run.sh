#!/bin/bash

while getopts "s:t:f:" arg; do
  case $arg in
    s) source=$OPTARG;;
    t) target=$OPTARG;;
    f) filetypes=$OPTARG;;
  esac
done

cd /go-media

app/builds/transfer -source "$source" -target "$target" -files "$filetypes"
app/builds/read
app/builds/transactions -file "config/commands.json"
app/builds/audit