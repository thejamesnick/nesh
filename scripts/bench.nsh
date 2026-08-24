# bench.nsh — quick perf sanity for this repo: build, then loop-race vs bash.
# Usage: nesh scripts/bench.nsh

try
  let code = run go build -o bin/nesh ./cmd/nesh
  if code != 0 then
    fail "build failed"
  end

  printf "%s\n" "==> nesh loop"
  bin/nesh benchmarks/nesh_scripts/loop.nsh > /dev/null

  printf "%s\n" "==> fixture pipeline"
  grep ERROR benchmarks/fixture.log | wc -l
on failure
  print "bench setup failed:" failure
  exit 1
end
print "done"
