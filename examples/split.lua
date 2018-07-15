local strings = require("strings")
local s = "123422\twowwowowow"
local rs = strings.split(s, "\t")

for k, v in pairs(rs) do
    print("value:("..v..")")
end
