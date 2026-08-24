#!/bin/sh
# redirect benchmark: write + 19 appends to a file, then read back a count
i=0
printf "line\n" > bench_out.tmp
while [ "$i" -lt 19 ]; do
  printf "line\n" >> bench_out.tmp
  i=$((i + 1))
done
wc -l < bench_out.tmp
