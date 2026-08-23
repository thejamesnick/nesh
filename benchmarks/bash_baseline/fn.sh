#!/usr/bin/env bash
# fn benchmark baseline: 500 function calls
add() {
  echo $(($1 + $2))
}

i=0; total=0
while [ $i -lt 500 ]; do
  total=$(add $total 2)
  i=$((i + 1))
done
echo "$total"
