# check.nsh — run before pushing: build, vet, test.
# Usage: nesh scripts/check.nsh

let failures = 0

try
  print "==> building"
  let code = run go build ./...
  if code != 0 then
    fail "build broken"
  end

  print "==> vetting"
  let code = run go vet ./...
  if code != 0 then
    fail "vet found issues"
  end

  print "==> testing"
  let code = run go test ./...
  if code != 0 then
    fail "tests failed"
  end
on failure
  print "CHECK FAILED:" failure
  let failures = 1
end

if failures == 0 then
  print "all green -- push away"
end
exit failures
