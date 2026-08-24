# redirect benchmark: write + 19 appends to a file, then read back a count
printf "line\n" > bench_out.tmp
let i = 1
while i < 20
  printf "line\n" >> bench_out.tmp
  let i = i + 1
end
wc -l < bench_out.tmp
