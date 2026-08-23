# Classic recursion checks
fn fact(n)
  if n <= 1 then
    return 1
  end
  return n * fact(n - 1)
end

fn fib(n)
  if n < 2 then
    return n
  end
  return fib(n - 1) + fib(n - 2)
end

print "5! =" fact(5)
print "fib(10) =" fib(10)

# iterative vs functional agree
let i = 1
let total = 1
while i <= 5
  let total = total * i
  let i = i + 1
end
print "iterative:" total
