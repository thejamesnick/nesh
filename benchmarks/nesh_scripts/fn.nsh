# fn benchmark: 500 non-recursive calls
fn add(a, b)
  return a + b
end

let i = 0
let total = 0
while i < 500
  let total = add(total, 2)
  let i = i + 1
end
print total
