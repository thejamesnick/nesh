#!/usr/bin/env bash
# vars benchmark baseline
x=0; y=41
while [ $x -lt 1000 ]; do
  z=$((y + 1))
  x=$((x + 1))
done
echo "$z"
