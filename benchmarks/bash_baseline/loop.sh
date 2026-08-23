#!/usr/bin/env bash
# loop benchmark baseline
i=0; total=0
while [ $i -lt 1000 ]; do
  total=$((total + i))
  i=$((i + 1))
done
echo "$total"
