# stdlib smoke test — strings, math, for-in
let csv = "deploy,verify,release"
let steps = split(csv, ",")

print len(steps) "steps"
for s in steps
  print "step:" upper(s)
end
print join(steps, " -> ")

print contains(csv, "verify")
print floor(2.7) round(3.4) abs(-5)
