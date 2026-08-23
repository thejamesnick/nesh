# String and number crunching — pure language, no system commands
let host = "web-01"
let up = true
let load = 0.75

if up and load < 1.0 then
  print "host" host "healthy, load" load
else
  print "investigate" host
end

fn status(ok)
  if not ok then
    return "down"
  end
  return "up"
end

print "status:" status(false)
print "negated:" not up
