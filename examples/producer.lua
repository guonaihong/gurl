local f = assert(io.open("test.file", "r"))

while true do
    local line = f:read("*line")
    if not line then break end
    out_ch:send(line)
end
