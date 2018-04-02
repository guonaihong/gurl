# gurl

#### 简介
gurl 是使用curl过程中的痛点的改进。gurl实现了本人经常使用的curl命令行选项。
日常使用curl脚本发现数据和行为耦合地太严重，如果支持配置文件就好了，如果配置文件里面支持变量，支持for，支持if，支持函数就好了。

#### 功能
* 定时运行gurl(支持cron表达式)
* 支持结构化配置文件(里面有if, else, for, func)

#### build
```
env GOPATH=`pwd` go get {github.com/NaihongGuo/flag,github.com/ghodss/yaml,github.com/robfig/cron}
env GOPATH=`pwd` go get github.com/satori/go.uuid
```
#### examples
* 命令行
  * ./gurl -F text="test" http://xxx.xxx.xxx.xxx:port
  * ./gurl -H "header1:value1" -H "header2:value2" http://xxx.xxx.xxx.xxx:port
* 配置文件  
 请见examples目录

#### TODO
* bugfix
* 一些用着很顺手的功能添加
