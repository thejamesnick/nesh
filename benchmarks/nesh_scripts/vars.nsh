# vars benchmark: repeated assignment and read
let x = 0
let y = 41
while x < 1000
  let z = y + 1
  let x = x + 1
end
print z
