# conn

## 简介
conn 是个tcp工具

## Install
``` bash
env GOPATH=`pwd` go get github.com/guonaihong/conn/conn
```

#### 命令行选项
```console
Usage of ./conn:
  -K string
        Read lua config from FILE
  -M, --mode string
        Set the server to read data from stdin or connect, and write to stdout or connect (default "c2c")
  -U    Use UNIX domain socket
  -ac int
        Number of multiple requests to make (default 1)
  -an int
        Number of requests to perform (default 1)
  -bench
        Run benchmarks test
  -conn-sleep string
        Sleep for a while before connecting
  -d, --data string
        Send data to the peer
  -duration string
        Duration of the test
  -k    Keep inbound sockets open for multiple connects
  -kargs string
        Command line parameters passed to the configuration file
  -l    Listen mode, for inbound connects
  -lt, --ltype string
        Add data header type
  -rate int
        Requests per second
  -send-rate string
        How many bytes of data in seconds
  -u    UDP mode
  -w string
        Timeout for connects and final net reads
  -z    Zero-I/O mode [used for scanning]

```
##### `-ac`
指定线程数, 开ac个线程, 发送an个请求

##### `-an`
指定次数

##### `-conn-sleep`
该选项只影响tcp 客户端模式下，先sleep -conn-sleep后面指定的时间，再connect

##### `-l`
变成一个简易的tcp服务服务端，后跟监听端口

##### `-lt`
该选项只在服务端模式下起作用, 写回的数据，会先加头
可选的值有u64be, u32be, u16be, 分别是64位，32位, 16位大端字节序

##### `-M`
控制tcp服务端把从connect读取的数据写到connect还是 stdint的数据写到终端
-M c2c 就是写回connect，变成一个简单的echo 服务，其它情况是stdin的数据写回到client

##### `-d --data`
写向对端的数据，支持@符号打开一个文件，目前仅支持客户端
```
conn -d "hello world" 127.0.0.1 1234
conn -d @./一个文件 127.0.0.1 1234
```

##### `-send-rate`
每隔一段时间发送多少字节数
支持单位符,ms(毫秒), s(秒), m(分), h(小时), d(天), w(周), M(月), y(年)
``` bash
conn -rate "8000B/250ms" 127.0.0.1 24986
```

##### `-w`
-w 指向connect 连接的超时或者 conn -l 监听模式下最长数据传输时间

##### `-z`
端口扫描
``` bash
conn -z 127.0.0.1 0-65532
```

###### `|`
管道模式，主要为了串联多个截然不同的conn命令，比如先启个服务端，再启个客户端连
``` bash
./conn -mode c2c -l 1234 "|" -conn-sleep 10ms -d "hello conn" 127.0.0.1 1234
```
lua script之间交换数据例子
``` bash
./conn -K producer.lua -kargs "-l ./test.list" "|" -K consumer.lua
```

##### `-K`
-K选项可以执行lua script，有关lua的用法，可以搜索下。
#### `-kargs`
该命令选选项主要从命令行传递参数给lua script

导入socket模块
```lua
local socket = require("socket")
```

创建tcp客户端并连接tcp server
```lua
local s = socket.new()
s:connect(addr)
```

写入数据到tcp(阻塞式)
```lua
s:write("hello world")
```

写数据，写成功返回，或者写超时返回
```lua
_, err = s:write(bytes, "1s")
if err ~= nil then
    print("write bytes", err, "\n")
end
```

对tcp读取4个字节数据
```lua
local len, err = s:read(4)
if err ~= nil then 
    print(err, "\n")
end
```

读4个字节，成功读取4个字节返回，或者超时返回错误
```lua
local len, err = s:read(4, "1s")
if err ~= nil then
    print(err, "\n")
end
```

关闭socket
```lua
s:close()
```
## TODO
* bugfix
* 顺手功能添加
