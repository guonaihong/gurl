# gurl
#### Documentation
* [English](./README_EN.md)

#### 简介
gurl 是http, websocket bench工具和curl的继承者

#### 功能
* 多协议支持http, websocket, tcp, udp
* 支持curl常用命令行选项
* 支持压测模式，可以根据并发数和线程数，也可以根据持续时间，也可以指定每秒并发数压测http, websocket, tcp服务
* 定时运行gurl(支持cron表达式)
* url 支持简写
* 支持管道模式

#### install
```bash
env GOPATH=`pwd` go get -u github.com/guonaihong/gurl/gurl
```

#### gurl 主命令选项
```console
Usage of gurl:
    http             Use the http subcommand
    tcp, udp         Use the tcp or udp subcommand
    ws, websocket    Use the websocket subcommand
```

#### websocket 子命用法
* [websocket](./wsurl/README.md)

#### tcp, udp 子命令用法
* [tcp udp](./conn/README.md)

#### http 子命令用法
```console
guonaihong https://github.com/guonaihong/gurl

Usage of gurl:
  -A, --user-agent string
    	Send User-Agent STRING to server (default "gurl")
  -F, --form string[]
    	Specify HTTP multipart POST data (H)
  -H, --header string[]
    	Pass custom header LINE to server (H)
  -I, --input-model
    	open input mode
  -J string[]
    	Turn key:value into {"key": "value"})
  -Jfa string[]
    	Specify HTTP multipart POST json data (H)
  -Jfa-string string[]
    	Specify HTTP multipart POST json data (H)
  -O, --output-mode
    	open output mode
  -R, --input-read string
    	open input file
  -W, --output-write string
    	open output file
  -X, --request string
    	Specify request command to use
  -ac int
    	Number of multiple requests to make (default 1)
  -an int
    	Number of requests to perform (default 1)
  -bench
    	Run benchmarks test
  -connect-timeout string
    	Maximum time allowed for connection
  -conns int
    	Max open idle connections per target host (default 10000)
  -cpus int
    	Number of CPUs to use
  -cron string
    	Cron expression
  -d, --data string
    	HTTP POST data
  -duration string
    	Duration of the test
  -form-string string[]
    	Specify HTTP multipart POST data (H)
  -input-fields string
    	sets the field separator (default " ")
  -l string
    	Listen mode, HTTP echo server
  -m, --merge
    	Combine the output results into the output
  -o, --output string
    	Write to FILE instead of stdout (default "stdout")
  -oflag string
    	Control the way you write(append|line|trunc)
  -r, --read-stream
    	Read data from the stream
  -rate int
    	Requests per second
  -skey, --input-setkey string
    	Set a new name for the default key
  -url string
    	Specify a URL to fetch
  -v, --verbose
    	Make the operation more talkative
  -w, --write-stream
    	Write data from the stream
  -wkey, --write-key string
    	Key that can be write

```
##### `-F 或 --form`
设置form表单, 比如-F text=文本内容，或者-F text=@./从文件里面读取, -F 选项的语义和curl命令一样
##### `--form-string`
和-F 或--form类似，不解释@符号，原样传递到服务端

##### `-ac`
指定线程数, 开ac个线程, 发送an个请求
```bash
gurl http -an 10 -ac 2 -F text=good :1234
```

##### `-an`
指定次数

##### `-bench`
压测模式，可以对http服务端进行压测，可以和-ac, -an, -duration, -rate 选项配合使用
 ``` console
    gurl http -bench -ac 25 -an 1000000 :1234
    Benchmarking 127.0.0.1 (be patient)
      Completed          100000 requests [2018-08-11 21:58:56.143]
      Completed          200000 requests [2018-08-11 21:59:00.374]
      Completed          300000 requests [2018-08-11 21:59:03.703]
      Completed          400000 requests [2018-08-11 21:59:06.559]
      Completed          500000 requests [2018-08-11 21:59:09.201]
      Completed          600000 requests [2018-08-11 21:59:11.757]
      Completed          700000 requests [2018-08-11 21:59:14.218]
      Completed          800000 requests [2018-08-11 21:59:16.639]
      Completed          900000 requests [2018-08-11 21:59:19.061]
      Completed         1000000 requests [2018-08-11 21:59:21.451]
      Finished          1000000 requests


    Server Software:        gurl-server
    Server Hostname:        
    Server Port:            1234

    Document Path:          
    Document Length:        0 bytes

    Status Codes:           200:1000000  [code:count]
    Concurrency Level:      10
    Time taken for tests:   28.807 seconds
    Complete requests:      1000000
    Failed requests:        0
    Total transferred:      137000000 bytes
    HTML transferred:       0 bytes
    Requests per second:    34713.37 [#/sec] (mean)
    Time per request:       0.288 [ms] (mean)
    Time per request:       0.029 [ms] (mean, across all concurrent requests)
    Transfer rate:          4755.73 [Kbytes/sec] received
    Percentage of the requests served within a certain time (ms)
      50%    0.21ms
      66%    0.31ms
      75%    0.38ms
      80%    0.42ms
      90%    0.57ms
      95%    0.66ms
      98%    0.79ms
      99%    0.89ms
     100%    16.45ms

 ```

##### `-duration`
和-bench选项一起使用，可以控制压测时间，支持单位符,ms(毫秒), s(秒), m(分), h(小时), d(天), w(周), M(月), y(年)
也可以混合使用 -duration 1m10s

##### `-connect-timeout`
设置http 连接超时时间。支持单位符,ms(毫秒), s(秒), m(分), h(小时), d(天), w(周), M(月), y(年)

##### `-rate`
指定每秒写多少条
``` console
gurl http -bench -ac 25 -an 3000 -rate 3000 :1234
Benchmarking 127.0.0.1 (be patient)
  Completed             300 requests [2018-08-11 22:02:01.625]
  Completed             600 requests [2018-08-11 22:02:01.725]
  Completed             900 requests [2018-08-11 22:02:01.825]
  Completed            1200 requests [2018-08-11 22:02:01.925]
  Completed            1500 requests [2018-08-11 22:02:02.025]
  Completed            1800 requests [2018-08-11 22:02:02.125]
  Completed            2100 requests [2018-08-11 22:02:02.225]
  Completed            2400 requests [2018-08-11 22:02:02.325]
  Completed            2700 requests [2018-08-11 22:02:02.425]
  Completed            3000 requests [2018-08-11 22:02:02.525]
  Finished             3000 requests


Server Software:        gurl-server
Server Hostname:        
Server Port:            1234

Document Path:          
Document Length:        0 bytes

Status Codes:           200:3000  [code:count]
Concurrency Level:      10
Time taken for tests:   1.000 seconds
Complete requests:      3000
Failed requests:        0
Total transferred:      411000 bytes
HTML transferred:       0 bytes
Requests per second:    3000.08 [#/sec] (mean)
Time per request:       3.333 [ms] (mean)
Time per request:       0.333 [ms] (mean, across all concurrent requests)
Transfer rate:          411.01 [Kbytes/sec] received
Percentage of the requests served within a certain time (ms)
  50%    0.17ms
  66%    0.18ms
  75%    0.18ms
  80%    0.19ms
  90%    0.21ms
  95%    0.23ms
  98%    0.26ms
  99%    0.31ms
 100%    1.34ms
```
##### `-d 或 --data`
发送http body数据到服务端, 支持@符号打开一个文件, 如果不接@直接把-d后面字符串发送到服务端
```bash
  gurl http -d "good" :12345
  gurl http -d "@./file" :12345
```

##### `-J`
-J 后面的key和value 会被组装成json字符串发送到服务端. key:value，其中value会被解释成字符串, key:=value，value会被解决成bool或者数字或者小数
  * 普通用法
```bash
  ./gurl http -J username:admin -J passwd:123456 -J bool_val:=true  -J int_val:=3 -J float_val:=0.3 http://127.0.0.1:12345
  {
    "bool_val": true,
    "float_val": 0.3,
    "int_val": 3,
    "passwd": "123456",
    "username": "admin"
  }
```
  * 嵌套用法
  ```bash
  ./gurl http -J a.b.c.d:=true -J a.b.c.e:=111 http://127.0.0.1:12345
  {
    "a": {
      "b": {
        "c": {
          "d:": true,
          "e:": 111
        }
      }
    }
  }
  ```

##### `-Jfa`
向multipart字段中插入json数据
```bash
./gurl http -Jfa text=DisplayText:good text=Language:cn text2=look:me -F text=good :12345

--4361c4e6ae1b083e9e0508a7b40eb215bccd265c4bed00137cc7d112e890
Content-Disposition: form-data; name="text"

{"DisplayText":"good","Language":"cn"}
--4361c4e6ae1b083e9e0508a7b40eb215bccd265c4bed00137cc7d112e890
Content-Disposition: form-data; name="text2"

{"look":"me"}
--4361c4e6ae1b083e9e0508a7b40eb215bccd265c4bed00137cc7d112e890
Content-Disposition: form-data; name="text"

good
--4361c4e6ae1b083e9e0508a7b40eb215bccd265c4bed00137cc7d112e890--
```
##### `-Jfa-string`
和-Jfa语法类似，不解析@符号

##### `-H 或者 --header`
设置http 头，可以指定多个
```bash
./gurl http -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
```

##### `-url`
设置http url的地址, 可以使用简写
* -url http://127.0.0.1:1234 --> 127.0.0.1:1234
* -url http://127.0.0.1:1234 --> :1234
* -url http://127.0.0.1/path --> /path

##### `-oflag`
-oflag 一般和-o选项配合使用(控制写文件的行为)
* -oflag append 默认-o的行为是新建文件然后写入，如果开启-ac -an选项，可以使用append肥所有的结果保存到一个文件中
* -oflag line 如果服务端返回的结果，想使用换行符分隔  
小提示: -oflag 后面的命令可以组合使用 "append|line"的意思是：把服务端的输出追加到某个文本中，并用'\n'分隔符


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

##### 批量访问多个url
开5个线程访问url.list里面的url列表
```
cat url.list

github.com
www.baidu.com
www.qq.com
www.taobao.com

gurl http -I -R url.list -skey "url=rf.col.0" "|" -ac 5 -r -w -merge "{url}" -o "/dev/null" "|" -O -wkey "status_code"
```

#### TODO
* 集群模式
* GUI
* tcp,udp 模块改造* tcp,udp 
