#!/bin/sh
# control flow benchmark: if/elif chain, 1000 iterations
i=0
hits=0
while [ "$i" -lt 1000 ]; do
  if [ "$i" -lt 250 ]; then
    hits=$((hits + 1))
  elif [ "$i" -lt 500 ]; then
    hits=$((hits + 2))
  elif [ "$i" -lt 750 ]; then
    hits=$((hits + 3))
  else
    hits=$((hits + 4))
  fi
  i=$((i + 1))
done
echo "$hits"
