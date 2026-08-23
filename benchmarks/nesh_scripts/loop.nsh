# loop benchmark: 1000 iterations of arithmetic
let i = 0
let total = 0
while i < 1000
  let total = total + i
  let i = i + 1
end
print total
