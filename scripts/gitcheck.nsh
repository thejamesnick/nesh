# gitcheck.nsh — quick repo glance: branch, dirtiness, recent history.
# Usage: nesh scripts/gitcheck.nsh

print "== branch =="
git rev-parse --abbrev-ref HEAD

print "== uncommitted changes =="
# NOTE: pipelines print to the screen today; `run` captures exit codes,
# not output (output capture is a tracked gap). So we just show status.
let code = run git status --porcelain
if code == 0 then
  print "(see list above if any)"
end

print "== last 3 commits =="
git --no-pager log --oneline -n 3
