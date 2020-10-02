#! /bin/bash

if [ -z "$REVIEWER" ]; then
  REVIEWER='auto'
fi

if [ -z "$1" ]; then
  # assume we are in a host directory
  if [ -f README.md ]; then
    sed '0,/+++/s//+++\nreviewer = "'"$REVIEWER"'"/' README.md | tee README.md.new
    mv README.md.new README.md
  fi
elif [ -f "hosts/$1/README.md" ]; then
  sed '0,/+++/s//+++\nreviewer = "'"$REVIEWER"'"/' "hosts/$1/README.md" | tee "hosts/$1/README.md".new
  mv "hosts/$1/README.md.new" "hosts/$1/README.md"
fi
