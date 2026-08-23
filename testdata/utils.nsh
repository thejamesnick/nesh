# shared helpers — imported by importer.nsh
let version = "1.0"

fn greet(name)
  return "hi " + name + " (from utils)"
end

fn shout(s)
  return upper(s) + "!"
end
