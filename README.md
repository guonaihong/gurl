# gurl

#### 简介
gurl 是使用curl过程中的痛点的改进。gurl实现了本人经常使用的curl命令行选项。
日常使用curl脚本发现数据和行为耦合地太严重，如果支持配置文件就好了，如果配置文件里面支持变量，支持for，支持if，支持函数就好了。

#### 功能
* 支持curl一部分功能
* 支持httpie一部功能
* 定时运行gurl(支持cron表达式)
* 支持结构化配置文件(里面有if, else, for, func)
* url 支持简写

#### build
```
env GOPATH=`pwd` go get {github.com/NaihongGuo/flag,github.com/ghodss/yaml,github.com/robfig/cron}
env GOPATH=`pwd` go get github.com/satori/go.uuid
env GOPATH=`pwd` go build gurl.go
```

#### examples
* 命令行
  * 发送multipart格式到服务端
  ```
  ./gurl -F text="test" http://xxx.xxx.xxx.xxx:port
  ```
  * 发送json格式数据到服务端
  ```
  ./gurl -J username:admin -J passwd:123456 -J bool_val:=true  -J int_val:=3 -J float_val:=0.3 http://127.0.0.1:12345
  {
    "bool_val": true,
    "float_val": 0.3,
    "int_val": 3,
    "passwd": "123456",
    "username": "admin"
  }

  ```
  * 如果key=value的数据是从文件或者终端里面读取，可以使用下面的方面转成json格式发给服务端
  ```
  echo "username=admin passwd=123456 bool_val:=true int_val:=3 float_val:=0.3"|xargs -d' ' -I {} echo -J {}|xargs ./gurl -url :12345
  {
    "bool_val": true,
    "float_val": 0.3,
    "int_val": 3,
    "passwd": "123456",
    "username": "admin"
  }
  ```
  * 发送多层json格式数据到服务端
  ```
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
  * multipart字段中有json数据格式
  ```
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
  * 指定多个http header
  ```
  ./gurl -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
  ```
  * 定时发送(每隔一秒从服务里取结果)
  ```
  ./gurl -cron "@every 1s" -H "session-id:f0c371f1-f418-477c-92d4-129c16c8e4d5" http://127.0.0.1:12345/asr/result
  ```
  * url支持简写种类
    * -url http://127.0.0.1:1234 --> 127.0.0.1:1234
    * -url http://127.0.0.1:1234 --> :1234
    * -url http://127.0.0.1/path --> /path
* 配置文件  
 请见examples目录
  ```
  * ./gurl -K ./demo.yml
  ```

#### TODO
* bugfix
* 一些用着很顺手的功能添加
