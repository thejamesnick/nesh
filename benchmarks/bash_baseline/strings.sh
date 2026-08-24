#!/bin/sh
# strings benchmark: 300 concatenations, then measure length
s=""
i=0
while [ "$i" -lt 300 ]; do
  s="${s}x"
  i=$((i + 1))
done
echo "${#s}"
