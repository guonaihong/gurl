# wsurl

#### 简介
websocket 压测工具

#### install
``` console
env GOPATH=`pwd` go build github.com/guonaihong/wsurl/wsurl
```

#### 命令行选项
```console
guonaihong https://github.com/guonaihong/wsurl

Usage of ./gurl:
  -A, --user-agent string
    	Send User-Agent STRING to server (default "gurl")
  -H, --header string[]
    	Pass custom header LINE to server (H) (default [])
  -I, --input-model
    	open input mode
  -O, --output-mode
    	open output mode
  -R, --input-read string
    	open input file
  -W, --output-write string
    	open output file
  -ac int
    	Number of multiple requests to make (default 1)
  -an int
    	Number of requests to perform (default 1)
  -bench
    	Run benchmarks test
  -binary
    	Send binary messages instead of utf-8
  -close
    	Send close message
  -duration string
    	Duration of the test
  -fsa, --first-send-after string
    	Wait for the first time before sending
  -input-fields string
    	sets the field separator (default " ")
  -l string
    	Listen mode, websocket echo server
  -ld, --last-packet string
    	The last packet is written to the connection
  -m, --merge
    	Combine the output results into the output
  -o, --output string
    	Write to FILE instead of stdout (default "stdout")
  -p, --packet string[]
    	Data packet to be send per connection
  -r, --read-stream
    	Read data from the stream
  -rate int
    	Requests per second
  -send-rate string
    	How many bytes of data in seconds
  -skey, --input-setkey string
    	Set a new name for the default key
  -url string
    	Specify a URL to fetch
  -w, --write-stream
    	Write data from the stream
  -wkey, --write-key string
    	Key that can be write

```

##### `-H 或header`
设置websocket 的header和http header类似

##### `-p 或 --packet`
发送websocket body数据到服务端，支持@符号打开一个文件, 如果不接@直接把-d后面字符串发送到服务端
```bash
  wsurl -p "good" :12345
  wsurl -p "@./file" :12345
```
##### `-send-rate`
``` bash
# 指定每多少ms发多少字节
wsurl -send-rate "8000B/250ms" -url ws://127.0.0.1:24986
```

##### `-binary`
默认是以text格式作为websocket消息类型, 加上-binary就以text作为消息类型

##### `-ld`
发送最后一个websocket包的内容
```bash
  wsurl -ld "good" :12345
  wsurl -ld "@./file" :12345
```

##### `-url`
设置websocket的url
* -url http://127.0.0.1:1234 --> 127.0.0.1:1234
* -url http://127.0.0.1:1234 --> :1234
* -url http://127.0.0.1/path --> /path

##### `-ac`
指定线程数, 开ac个线程, 发送an个请求
```bash
wsurl -an 10 -ac 2 -F text=good :1234
```

##### `-an`
指定次数

##### `-duration`
和-bench选项一起使用，可以控制压测时间，支持单位符,s(秒), m(分), h(小时), d(天), w(周), M(月), y(年), ms(毫秒)
也可以混合使用 -duration 1m10s

##### `-rate`
指定每秒写多少条，目前只有打开-bench选项才起作用

##### `-close`
客户端主动发起close消息给服务端

##### `-bench`
压测模式
wsurl -bench -ac 20 -an 10000 -url :33333 -close
``` console
Connecting to to ws://127.0.0.1:33333
    Opened            1000 connections: [2018-08-23 20:50:55.987]
    Opened            2000 connections: [2018-08-23 20:50:56.129]
    Opened            3000 connections: [2018-08-23 20:50:56.266]
    Opened            4000 connections: [2018-08-23 20:50:56.409]
    Opened            5000 connections: [2018-08-23 20:50:56.552]
    Opened            6000 connections: [2018-08-23 20:50:56.684]
    Opened            7000 connections: [2018-08-23 20:50:56.835]
    Opened            8000 connections: [2018-08-23 20:50:56.098]
    Opened            9000 connections: [2018-08-23 20:50:57.125]
    Opened           10000 connections: [2018-08-23 20:50:57.268]

    Finished 10000 connections

Concurrency Level:        20
Time taken for tests:     1.432677765s
Connected:                10000
Disconnected:             0
Failed:                   0
Total transferred:        0
Total received            0
Requests per second:      6979 [#/sec] (mean)
Time per request:         716338.883 [ms] (mean)
Time per request:         71.634 [ms] (mean, across all concurrent requests)
Transfer rate:            0.000 [Kbytes/sec] received

Percentage of the requests served within a certain time (ms)
    50%    2.00ms
    66%    2.00ms
    75%    3.00ms
    80%    3.00ms
    90%    4.00ms
    95%    6.00ms
    98%    7.00ms
    99%    9.00ms
    100%   21.00ms
```

#### 高级主题(stream功能)
##### `-I`
打开input模式

##### `-R`
打开列表文件, 可以使用-input-fields 指定分割符，默认是空格

##### `-skey`
给默认的名字取个别名，相当于取个好听的变量名，方便后面引用

##### `-r`
从流里面读取数据

##### `-w`
结果输出到流

##### `-merge`
把输入流里面的结果和识别结果组成大的结果，写到输出流

##### `-O`
打开output模式

##### `-wkey`
控制写出的json key

##### `|`
管道符主要拼接多个gurl功能块
