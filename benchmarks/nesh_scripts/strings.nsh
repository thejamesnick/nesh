# strings benchmark: 300 concatenations, then measure
let s = ""
let i = 0
while i < 300
  let s = s + "x"
  let i = i + 1
end
print len(s)
