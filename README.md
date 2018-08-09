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
* 支持ab统计报表，并且性能比ab命令更高，每秒可以压测更多消息数
* 定时运行gurl(支持cron表达式)
* 支持lua语言作为配置文件(可以写if, else, for, func)
* url 支持简写
* 支持管道模式

#### 推荐的使用流程
1. 使用gurl命令行选项调通功能，很多用户经过这一步就完成了任务
1. 如果命令行里面有难以记忆的部分，并且也需要经常使用它，可使用gurl -gen &>demo.lua 命令保存到文件里
2. gurl -K demo.lua根据配置文件里面的数据访问服务端
3. 如果要修改demo.lua，也不熟悉配置文件的写法，可以通过./gurl -K demo.lua -gen 命令把demo.lua的数据转成命令行，然后重复第1步操作

#### install
```bash
env GOPATH=`pwd` go get -u github.com/guonaihong/gurl
```

#### examples
* 命令行
 * bench
 使用-bench选项打开bench(压测)模式，可以压测服务端，诊断性能瓶颈
 ac可以指定线程数, an可以指定跑多少次 ，输出如下 
 ```bash
 ./gurl -ac 10 -an 1000000 -F text=good  -bench http://127.0.0.1:12346 
 
    Benchmarking 127.0.0.1 (be patient)
    Completed 100000 requests
    Completed 200000 requests
    Completed 300000 requests
    Completed 400000 requests
    Completed 500000 requests
    Completed 600000 requests
    Completed 700000 requests
    Completed 800000 requests
    Completed 900000 requests
    Completed 1000000 requests
    Finished 1000000 requests
    
    
    Server Software:        gnc
    Server Hostname:        
    Server Port:            12346
    
    Document Path:          
    Document Length:        0 bytes
    
    Concurrency Level:      10
    Time taken for tests:   35.293 seconds
    Complete requests:      1000000
    Failed requests:        0
    Total transferred:      131000000 bytes
    HTML transferred:       0 bytes
    Requests per second:    28334.54 [#/sec] (mean)
    Time per request:       0.353 [ms] (mean)
    Time per request:       0.035 [ms] (mean, across all concurrent requests)
    Transfer rate:          3711.83 [Kbytes/sec] received
    Percentage of the requests served within a certain time (ms)
      50%    0
      66%    0
      75%    0
      80%    0
      90%    0
      95%    0
      98%    1
      99%    1
     100%    14

 ```
 使用-bench 选项打开bench(压测)模式, -rate 指定每秒写多少条消息
```
    ./gurl -bench -ac 10 -an 3000 -rate 3000 :1234
    Benchmarking 127.0.0.1 (be patient)
    Completed     300 requests [2018-08-09 21:43:20.643]
    Completed     600 requests [2018-08-09 21:43:20.743]
    Completed     900 requests [2018-08-09 21:43:20.843]
    Completed    1200 requests [2018-08-09 21:43:20.943]
    Completed    1500 requests [2018-08-09 21:43:21.043]
    Completed    1800 requests [2018-08-09 21:43:21.143]
    Completed    2100 requests [2018-08-09 21:43:21.243]
    Completed    2400 requests [2018-08-09 21:43:21.343]
    Completed    2700 requests [2018-08-09 21:43:21.443]
    Completed    3000 requests [2018-08-09 21:43:21.543]
    Finished 3000 requests


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
    Requests per second:    2999.41 [#/sec] (mean)
    Time per request:       3.334 [ms] (mean)
    Time per request:       0.333 [ms] (mean, across all concurrent requests)
    Transfer rate:          410.92 [Kbytes/sec] received
    Percentage of the requests served within a certain time (ms)
      50%    0
      66%    0
      75%    0
      80%    0
      90%    0
      95%    0
      98%    0
      99%    0
     100%    7

```
  * 管道模式
  ```bash
 ./gurl -an 1 -K ./producer.lua -kargs "-l all.txt" "|" -an 0 -ac 12 -K ./http_slice.lua -kargs "-appkey xx -url http://192.168.6.128:24990/asr/pcm " "|" -an 0 -K ./write_file.lua -kargs "-f asr.result"
  ```
  * 发送multipart格式到服务端
  ```bash
  # 1.发送字符串test到服务端
  ./gurl -F text="test" http://xxx.xxx.xxx.xxx:port
  # 2.打开名为file的文件，并用其内容发送到服务端
  ./gurl -F text="@./file" http://xxx.xxx.xxx.xxx:port
  ```
  * 发送http body数据到服务端
  ```bash
  # 1.发送字符串test到服务端
  gurl -d "good" :12345
  # 2.打开为file的文件，并用其内容发送到服务端
  gurl -d "@./file" :12345
  ```
  * 发送json格式数据到服务端
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
  * 如果key:value的数据是从文件或者终端里面读取，可以使用下面的方面转成json格式发给服务端
  ```bash
  echo "username:admin passwd:123456 bool_val:=true int_val:=3 float_val:=0.3"|xargs -d' ' -I {} echo -J {}|xargs ./gurl -url :12345
  {
    "bool_val": true,
    "float_val": 0.3,
    "int_val": 3,
    "passwd": "123456",
    "username": "admin"
  }
  ```
  * 发送多层json格式数据到服务端
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
  * 向multipart字段中插入json数据
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
  * 开ac个线程, 发送an个请求
  ```bash
  ./gurl -an 10 -ac 2 -F text=good :1234
  ```
  * 指定多个http header
  ```bash
  ./gurl -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
  ```
  * 定时发送(每隔一秒从服务里取结果)
  ```bash
  ./gurl -cron "@every 1s" -H "session-id:f0c371f1-f418-477c-92d4-129c16c8e4d5" http://127.0.0.1:12345/asr/result
  ```
  * url支持简写种类
    * -url http://127.0.0.1:1234 --> 127.0.0.1:1234
    * -url http://127.0.0.1:1234 --> :1234
    * -url http://127.0.0.1/path --> /path

  * -oflag 一般和-o选项配合使用(控制写文件的行为)
    * -oflag append 默认-o的行为是新建文件然后写入，如果开启-ac -an选项，可以使用append肥所有的结果保存到一个文件中
    * -oflag line 如果服务端返回的结果，想使用换行符分隔  
    小提示: -oflag 后面的命令可以组合使用 "append|line"的意思是：把服务端的输出追加到某个文本中，并用'\n'分隔符
 * 配置文件
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
#### 高级用法
高级用法主要讲如何使用gurl内置的lua函数，以下代码都可以通过-K 选项执行，-karg "这里是从给脚本的命令行参数"
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
    time.sleep(250, "ms")
    time.sleep(1, "s")
    time.sleep(1, "m")
    time.sleep(1, "h")
```
#### TODO
* bugfix
* 一些用着很顺手的功能添加
