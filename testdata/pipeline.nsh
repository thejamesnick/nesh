# Deploy pipeline simulation — control flow, functions, exit codes via run
fn check(name)
  if name == "prod" then
    return true
  end
  return false
end

let env = "staging"
if check(env) then
  print "extra care for prod"
end

let code = run deploy-staging --dry-run
print "deploy exited" code

fn retry(n, msg)
  let attempt = 1
  while attempt <= n
    print "attempt" attempt ":" msg
    let attempt = attempt + 1
  end
  return attempt > n
end

retry(2, "pushing artifacts")
print "pipeline done"
