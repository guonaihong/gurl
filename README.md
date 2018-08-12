# gurl
#### Documentation
* [English](./README_EN.md)

#### 简介
gurl 是使用curl过程中的痛点改进。gurl实现了本人经常使用的curl命令行选项。
日常使用curl脚本发现数据和行为耦合地太严重，如果支持配置文件就好了，如果配置文件里面支持变量，支持for，支持if，支持函数就好了。
支持在命令行上方便写json数据就好了。
最后说下gurl的哲学， gurl希望做一个安静的工具，不希望有难用的功能，让使用者爽完就离开它。

#### 功能
* 支持curl常用命令行选项
* 支持压测模式，可以根据并发数和线程数，也可以根据持续时间，也可以指定每秒并发数压测http服务
* 定时运行gurl(支持cron表达式)
* 支持lua语言作为配置文件(可以写if, else, for, func)
* url 支持简写
* 支持管道模式

#### install
```bash
env GOPATH=`pwd` go get -u github.com/guonaihong/gurl
```

#### 命令行选项
```console
Usage of ./gurl:
  -A, --user-agent string
        Send User-Agent STRING to server (default "gurl")
  -F, --form string[]
        Specify HTTP multipart POST data (H)
  -H, --header string[]
        Pass custom header LINE to server (H)
  -J string[]
        Turn key:value into {"key": "value"})
  -Jfa string[]
        Specify HTTP multipart POST json data (H)
  -K, --config string
        lua script
  -X, --request string
        Specify request command to use
  -ac int
        Number of multiple requests to make (default 1)
  -an int
        Number of requests to perform (default 1)
  -bench
        Run benchmarks test
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
  -echo string
        HTTP echo server
  -gen
        Generate the default lua script
  -kargs string
        Command line parameters passed to the configuration file
  -o, --output string
        Write to FILE instead of stdout (default "stdout")
  -oflag string
        Control the way you write(append|line|trunc)
  -rate int
        Requests per second
  -url string
        Specify a URL to fetch
  -v, --verbose
        Make the operation more talkative
```
##### `-F 或 --form`
设置form表单, 比如-F text=文本内容，或者-F text=@./从文件里面读取, -F 选项的语义和curl命令一样

##### `-ac`
指定线程数, 开ac个线程, 发送an个请求
```bash
./gurl -an 10 -ac 2 -F text=good :1234
```

##### `-an`
指定次数

##### `-bench`
压测模式，可以对http服务端进行压测，可以和-ac, -an, -duration, -rate 选项配合使用
 ```bash
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
和-bench选项一起使用，可以控制压测时间，支持单位符,s(秒), m(分), h(小时), d(天), w(周), M(月), y(年)  
也可以混合使用 -duration 1m10s

##### `-rate`
指定每秒写多少条，目前只有打开-bench选项才起作用
```
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
##### `|`
管道模式, 主要为了串联多个lua脚本而设计，第1个脚本的输出，变成第2个脚本的输入
```bash
 ./gurl -an 1 -K ./producer.lua -kargs "-l all.txt" "|" -an 0 -ac 12 -K ./http_slice.lua -kargs "-appkey xx -url http://192.168.6.128:24990/asr/pcm " "|" -an 0 -K ./write_file.lua -kargs "-f asr.result"
```

##### `-d 或 --data`
发送http body数据到服务端, 支持@符号打开一个文件, 如果不接@直接把-d后面字符串发送到服务端
```bash
  gurl -d "good" :12345
  gurl -d "@./file" :12345
```

##### `-J`
-J 后面的key和value 会被组装成json字符串发送到服务端. key:value，其中value会被解释成字符串, key:=value，value会被解决成bool或者数字或者小数
  * 普通用法
```bash
  ./gurl -J username:admin -J passwd:123456 -J bool_val:=true  -J int_val:=3 -J float_val:=0.3 http://127.0.0.1:12345
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
  ./gurl -J a.b.c.d:=true -J a.b.c.e:=111 http://127.0.0.1:12345
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
./gurl -Jfa text=DisplayText:good -Jfa text=Language:cn -Jfa text2=look:me -F text=good :12345

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

##### `-H 或者 --header`
设置http 头，可以指定多个
```bash
./gurl -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
```

##### `-cron`
定时发送(每隔一秒从服务里取结果)
```bash
./gurl -cron "@every 1s" -H "session-id:f0c371f1-f418-477c-92d4-129c16c8e4d5" http://127.0.0.1:12345/asr/result
```

##### `-url`
* -url http://127.0.0.1:1234 --> 127.0.0.1:1234
* -url http://127.0.0.1:1234 --> :1234
* -url http://127.0.0.1/path --> /path


##### `-oflag`
-oflag 一般和-o选项配合使用(控制写文件的行为)
* -oflag append 默认-o的行为是新建文件然后写入，如果开启-ac -an选项，可以使用append肥所有的结果保存到一个文件中
* -oflag line 如果服务端返回的结果，想使用换行符分隔  
小提示: -oflag 后面的命令可以组合使用 "append|line"的意思是：把服务端的输出追加到某个文本中，并用'\n'分隔符

##### `-gen`
* 从命令行的数据生成配置文件(选项 -gen)
```lua
./gurl -X POST -F mode=A -F text=good -F voice=@./good.opus -url http://127.0.0.1:24909/eval/opus -gen &>demo.lua 

#todo

```
* 把配置文件转成命令行形式(选项-gen -K 配置文件)
```bash
./gurl -K demo.lua -gen
gurl -X POST -F mode=A -F text=good -F voice=@./good.opus -url http://127.0.0.1:24909/eval/opus
```

#### `-K`
-K选项可以执行lua script，有关lua的用法，可以搜索下。

#### `-kargs`
该命令选选项主要从命令行传递参数给lua script

下而的example讲如何使用gurl内置的lua函数，以下代码都可以通过-K 选项执行，-kargs "这里是从给脚本的命令行参数"
* 在配置文件里面解析命令行配置
```lua
    local cmd = require("cmd")
    local flag = cmd.new()
    local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("a, addr", "", "Remote service address")
            :parse("-f ./tst.pcm -a 127.0.0.1:8080")

    function tableHasKey(table, key)
        return table[key] ~= nil 
    end

    if (not tableHasKey(opt, "f")) or
        (not tableHasKey(opt, "file")) or
        (not tableHasKey(opt, "a")) or
        (not tableHasKey(opt, "addr")) then

        opt.Usage()

        return
    end

    for k, v in pairs(opt) do
        print("cmd opt ("..k..") parse value ("..v..")")
    end

```
* 发送http请求
```lua
    local http = require("http")
    local rsp = http.send({
        H = { 
            "appkey:"..config.appkey,
            "X-Number:"..xnumber,
            "session-id:"..session_id,
        },
        MF = {
            "voice=" .. bytes,
        },
        url = config.url
    })

    --print("bytes ("..bytes..")")
    if #rsp["err"] ~= 0 then
        print("rsp error is ".. rsp["err"])
        return
    end

    if rsp["status_code"] == 200 then
        body = rsp["body"]
        if #rsp["body"] == 0 then
             body = "{}"
        end
        print(json.format(body))
    else
        print("error http code".. rsp["status_code"])
    end

```

* sleep
```lua
    local time = require("time")
    time.sleep("250ms")
    time.sleep("1s")
    time.sleep("1m")
    time.sleep("1h")
    time.sleep("1s250ms")
```
#### TODO
* bugfix
* 一些用着很顺手的功能添加
