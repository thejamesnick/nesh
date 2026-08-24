# control flow benchmark: if/elif chain, 1000 iterations
let i = 0
let hits = 0
while i < 1000
  if i < 250 then
    let hits = hits + 1
  elif i < 500 then
    let hits = hits + 2
  elif i < 750 then
    let hits = hits + 3
  else
    let hits = hits + 4
  end
  let i = i + 1
end
print hits
