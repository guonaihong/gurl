# gurl

#### Documentation
* [chinese](./README.md)

#### Introduction
Gurl is the successor to http, websocket bench tools and curl

#### Features
* Multi-protocol support http, websocket, tcp, udp
* Supports some of curl's features
* Supports some of the functions of ab, and the performance is higher than the ab command
* Support regular running gurl (support cron expression function)
* Support lua as configuration file (support for if, else, for, func)
* Url support abbreviations
* Support pipeline mode

#### install
```bash
env GOPATH=`pwd` go get -u github.com/guonaihong/gurl/gurl
```
#### gurl main command usage
```console
Usage of gurl:
    http             Use the http subcommand
    tcp, udp         Use the tcp or udp subcommand
    ws, websocket    Use the websocket subcommand
```
#### websocket subcommand usage
* [websocket](./wsurl/README.md)

#### tcp, udp subcommand usage
* [tcp udp](./conn/README.md)

#### http subcommand usage
```console
guonaihong https://github.com/guonaihong/gurl

Usage of gurl:
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
  -Jfa-string string[]
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
  -gen
        Generate the default lua script
  -kargs string
        Command line parameters passed to the configuration file
  -l string
        Listen mode, HTTP echo server
  -o, --output string
        Write to FILE instead of stdout (default "stdout")
  -oflag string
        Control the way you write(append|line|trunc)
  -rate int
        Requests per second
  -url string
        Specify a URL to fetch
```

##### `-F 或 --form`
Set the form form, such as -F text=text content, or -F text=@./read from the file, the semantics of the -F option are the same as the curl command
##### `--form-string`
Similar to -F or --form, without interpreting the @symbol, passed to the server as it is

##### `-ac`
Specify the number of threads, open ac threads, send an request
```bash
./gurl http -an 10 -ac 2 -F text=good :1234
```

##### `-an`
Specified number

##### `-bench`
In the pressure test mode, the http server can be pressed and used with the -ac, -an, -duration, -rate option.
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
Used with the -bench option to control the press time, support unit, s (seconds), m (minutes), h (hours), d (days), w (weeks), M (months), y (years) )
Can also be mixed -duration 1m10s

##### `-connect-timeout`
Set the http connection timeout period. Support unit, ms (milliseconds), s (seconds), m (minutes), h (hours), d (days), w (weeks), M (months), y (years)

##### `-rate`
Specify how many times to write per second
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
##### `|`
Pipeline mode, mainly designed to concatenate multiple lua scripts, the output of the first script becomes the input of the second script
```bash
 ./gurl http -an 1 -K ./producer.lua -kargs "-l all.txt" "|" -an 0 -ac 12 -K ./http_slice.lua -kargs "-appkey xx -url http://192.168.6.128:24990/asr/pcm " "|" -an 0 -K ./write_file.lua -kargs "-f asr.result"
```

##### `-d 或 --data`
Send http body data to the server, support @symbol to open a file, if not @ directly send the string after -d to the server
```bash
  gurl http -d "good" :12345
  gurl http -d "@./file" :12345
```

##### `-J`
The key and value after -J will be assembled into a json string and sent to the server. key:value, where value will be interpreted as a string, key:=value, value will be resolved to bool or number or decimal
  * Common usage
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
  * Nested usage
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
Insert json data into the multipart field
```bash
./gurl http -Jfa text=DisplayText:good -Jfa text=Language:cn -Jfa text2=look:me -F text=good :12345

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
Similar to the -Jfa syntax, does not parse the @symbol

##### `-H 或者 --header`
Set http headers, you can specify multiple
```bash
./gurl http -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
```

##### `-cron`
Send regularly (take results from the service every second)
```bash
./gurl http -cron "@every 1s" -H "session-id:f0c371f1-f418-477c-92d4-129c16c8e4d5" http://127.0.0.1:12345/asr/result
```

##### `-url`
* -url http://127.0.0.1:1234 --> 127.0.0.1:1234
* -url http://127.0.0.1:1234 --> :1234
* -url http://127.0.0.1/path --> /path


##### `-oflag`
-oflag Generally used in conjunction with the -o option (controls the behavior of writing files)
* -oflag append The default -o behavior is to create a new file and then write it. If you enable the -ac -an option, you can use append to save all the results to a file.
* -oflag line If the server returns the result, I want to use a newline to separate
Tip: The commands after -oflag can be combined with "append|line" to mean appending the output of the server to a text with the '\n' separator

##### `-gen`
* Generate a configuration file from the command line's data (option -gen)
```lua
gurl http -X POST -F mode=A -F text=good -F voice=@./good.opus -url http://127.0.0.1:24909/eval/opus -gen &>demo.lua 

#todo

```
* Convert the configuration file to the command line format (option -gen -K configuration file)
```bash
gurl http -K demo.lua -gen
gurl http -X POST -F mode=A -F text=good -F voice=@./good.opus -url http://127.0.0.1:24909/eval/opus
```

#### `-K`
The -K option can execute lua script. For the usage of lua, you can search for it.

#### `-kargs`
This command option mainly passes parameters from the command line to lua script.

The following example shows how to use gurl built-in lua function. The following code can be executed with the -K option, -kargs "here is the command line argument from the script"
* Parse the command line configuration in the configuration file
```lua
    local flag = require("flag").new()
    local opt = flag
                :opt_str("f, file", "", "open audio file")
                :opt_str("a, addr", "", "Remote service address")
                :parse("-f ./tst.pcm -a 127.0.0.1:8080")


    if  #opt["f"] == 0 or
        #opt["file"] == 0 or
        #opt["a"] == 0 or
        #opt["addr"] == 0  then

        opt.Usage()

        return
    end

    for k, v in pairs(opt) do
        print("cmd opt ("..k..") parse value ("..v..")")
    end
```
* Send http request
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
* Some add with very handy features
