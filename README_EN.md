# gurl

#### Documentation
* [chinese](./README.md)

#### Introduction
Gurl is a pain point in the process of using curl.Gurl implements curl command line options.In daily use, it is found that the data and behavior of the curl script are severely coupled. If the configuration file is supported and the configuration file supports variables, the for, if, and function functions are supported. Support for writing json data on the command line is just fine.Finally, under the philosophy of gurl, gurl hopes to be a quiet tool. It doesn't want to be called bad, let users run away and leave it.

#### Features
- Supports some of curl's features
- Supports some of the features of httpie
- Supports some of the functions of ab, and the performance is higher than the ab command
- Support regular running gurl (support cron expression function)
- Support js as configuration file (support for if, else, for, func)
- Url support abbreviations

#### Recommended use process
1、Using the gurl command line option to enable the function, many users complete the task after this step.
2、If there is a hard-to-remember part of the command line and you need to use it often, save it to a file using the gurl -gen &>demo.cf.js command.
3、gurl -K demo.cf.js accesses the server based on the data in the configuration file.
4、If you want to modify demo.cf.js and you are not familiar with the configuration of the configuration file, you can use the ./gurl -K demo.cf.js -gen command to convert the demo.cf.js data to the command line, and then repeat step 1. operating.

#### install
```bash
env GOPATH=`pwd` go get -u github.com/guonaihong/gurl
```

#### examples
* Command Line
  * Send multipart format to server
  ```bash
  # 1.Send the string “test” to the server
  ./gurl -F text="test" http://xxx.xxx.xxx.xxx:port
  # 2.Open the file named “file” and send it to the server with its contents
  ./gurl -F text="@./file" http://xxx.xxx.xxx.xxx:port
  ```
  * Send http body data to server
  ```bash
  # 1.Send the string “test” to the server
  gurl -d "good" :12345
  # 2.Open the file named “file” and send it to the server with its contents
  gurl -d "@./file" :12345
  ```
  * Send json format data to server
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
  * If the data of "key:value" is read from the file or the terminal, you can use the following aspects to send it to the server in json format.
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
  * Send multi-layer json format data to server
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
  * Insert json data into the multipart field

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
  * Open "ac" threads, send "an" request

  ```bash
  ./gurl -an 10 -ac 2 -F text=good :1234
  ```
  * Specify multiple http headers

  ```bash
  ./gurl -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
  ```
  * Send regularly (every second from the service)

  ```bash
  ./gurl -cron "@every 1s" -H "session-id:f0c371f1-f418-477c-92d4-129c16c8e4d5" http://127.0.0.1:12345/asr/result
  ```
  * Url supports abbreviated types
    * -url http://127.0.0.1:1234 --> 127.0.0.1:1234
    * -url http://127.0.0.1:1234 --> :1234
    * -url http://127.0.0.1/path --> /path

  * "-oflag" is generally used with the "-o" option (controls the behavior of writing files)

    * -oflag append: the default behavior of "-o" is to create a new file and then write it. If the "-ac -an" option is enabled, you can use append to save all the results in a file.
    * -oflag line: if the server returns a result, you want to use a line break to separate.  
    Tips: Commands following "-oflag" can be combined,
     "append|line" mean：Append the server's output to a text and use '\n' separator.
 * Profile
   * Generate a configuration file from the command line data (option -gen)

  ```js
  ./gurl -X POST -F mode=A -F text=good -F voice=@./good.opus -url http://127.0.0.1:24909/eval/opus -gen &>demo.cf.js 
  var cmd = {
      "method": "POST",
      "F": [
          "mode=A",
          "text=good",
          "voice=@./good.opus"
      ],
      "url": "http://127.0.0.1:24909/eval/opus"
  };
  var rsp = gurl_send(cmd);
  console.log(rsp.body);
  ```
  * Turn the configuration file into a command line (option -gen -K configuration file)
  ```bash
  ./gurl -K demo.cf.js -gen
  gurl -X POST -F mode=A -F text=good -F voice=@./good.opus -url http://127.0.0.1:24909/eval/opus
  ```
 * bench  
 Use the "-bench" option to turn on the benchmark mode to test the server and diagnose performance bottlenecks. The output is as follows

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
#### TODO
* bugfix
* Add some very handy features
