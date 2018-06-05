#### Introduction
"Gurl-bench" mode and "ab" command horizontal comparison evaluation.

#### Documentation
* [English](./gurl_vs_ab.md)

* Ready
``` bash
# Start the "http echo" echo service provided by gurl

gurl -echo :12345
```
* gurl
```
gurl -ac 21 -an 1000000 -bench :12345

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


Server Software:        gurl-server
Server Hostname:        
Server Port:            12345

Document Path:          
Document Length:        0 bytes

Concurrency Level:      21
Time taken for tests:   7.708 seconds
Complete requests:      1000000
Failed requests:        0
Total transferred:      137000000 bytes
HTML transferred:       0 bytes
Requests per second:    129741.42 [#/sec] (mean)
Time per request:       0.162 [ms] (mean)
Time per request:       0.008 [ms] (mean, across all concurrent requests)
Transfer rate:          17774.57 [Kbytes/sec] received
Percentage of the requests served within a certain time (ms)
  50%    0
  66%    0
  75%    0
  80%    0
  90%    0
  95%    0
  98%    0
  99%    0
 100%    40

```

* ab
```
This is ApacheBench, Version 2.3 <$Revision: 1706008 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.01 (be patient)
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


Server Software:        gurl-server
Server Hostname:        127.0.01
Server Port:            12345

Document Path:          /
Document Length:        0 bytes

Concurrency Level:      21
Time taken for tests:   33.300 seconds
Complete requests:      1000000
Failed requests:        0
Total transferred:      137000000 bytes
HTML transferred:       0 bytes
Requests per second:    30029.76 [#/sec] (mean)
Time per request:       0.699 [ms] (mean)
Time per request:       0.033 [ms] (mean, across all concurrent requests)
Transfer rate:          4017.65 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.1      0       3
Processing:     0    0   0.4      0     212
Waiting:        0    0   0.4      0     212
Total:          0    1   0.4      1     212

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      1
  99%      1
 100%    212 (longest request)
```

* In conclusion
  
The gurl command can send more messages per second (12w/s) than the ab(3w/s) command. The more cores, the higher the performance of "gurl" than "ab".The reason "gurl" is faster than "ab" is that ab only uses a single thread for pressure testing, and is particularly dependent on the cpu clock speed. The fast clock speed runs faster.
