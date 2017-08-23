# LoadBalancer Benchmark

## 环境准备

编写了一个小程序做 Backend Service 对外提供一个 api，模拟 api 处理时间为 **200ms** （sleep 200ms）

搭建的环境参数如下：

| 组件                        | 副本数量 | CPU(m, 1 core = 1000m) | memory(MiB) |
| ------------------------- | ---- | ---------------------- | ----------- |
| ingress controller(nginx) | 1    | 1000                   | 1024        |
| backend pod               | 1    | 1000                   | 1024        |



## 测试结果

使用 `ab` 进行测试

在 1000 Concurrency 下

| type      | nginx CPU/Memory | backend pod CPU/Memory | r/s     | wait time(ms) /r | server process time (ms) / r |
| --------- | ---------------- | ---------------------- | ------- | ---------------- | ---------------------------- |
| NodePort  | -                | 900m/67MiB             | 3747.75 | 266.826          | 0.267                        |
| LB - TCP  | 680m/130MiB      | 950m/98MiB             | 3867.60 | 258.558          | 0.259                        |
| LB - HTTP | 854m/130MiB      | 890m/79Mi              | 1558.15 | 641.787          | 0.642                        |



在 200 Concurrency 下

| type      | nginx CPU/Memory | backend pod CPU/Memory | r/s     | wait time(ms) /r | server process time (ms) / r |
| --------- | ---------------- | ---------------------- | ------- | ---------------- | ---------------------------- |
| NodePort  | -                | 900m/67MiB             | 3463.86 | 577.391          | 0.289                        |
| LB - TCP  | 544m/135MiB      | 977m/88MiB             | 3153.10 | 634.296          | 0.317                        |
| LB - HTTP | 887m/136MiB      | 826m/64MiB             | 1525.92 | 1310.685         | 0.655                        |



## 总结

1.  可以看出 LoadBalancer 在四层负载下的性能与 NodePort 基本持平（理论上略有下降， 因为会先经过 LVS，再经过 NGINX）在七层负载下，比 NodePort 和四层负载损失了约 **50%-60%** 的性能。
2.  在高并发下，对 Ingress Controller CPU 的压力比较大，内存倒不是很吃紧
3.  可以看出 Backend Pod 的 CPU 基本跑满了，而 Nginx 还有闲余，在这个测试中，是 Backend Pod 限制了 rps 的上限，对 Backend 进行 2 倍扩容后理论上能达到 2 倍的 rps
4.  七层的负载比四层的负载更加消耗 ingress controller CPU 的资源，因为在七层负载中，nginx 需要去解析 HTTP 的包，会需要更多的 CPU
5.  Ingress Controller 的资源配置建议是 `1 Core / 300 MiB` 的比例

## 详细报告

```
Server Software:
Server Hostname:        192.168.18.213
Server Port:            32555

Document Path:          /bench200
Document Length:        18 bytes

Concurrency Level:      1000
Time taken for tests:   26.683 seconds
Complete requests:      100000
Failed requests:        0
Write errors:           0
Keep-Alive requests:    100000
Total transferred:      16500000 bytes
HTML transferred:       1800000 bytes
Requests per second:    3747.75 [#/sec] (mean)
Time per request:       266.826 [ms] (mean)
Time per request:       0.267 [ms] (mean, across all concurrent requests)
Transfer rate:          603.89 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1  11.0      0     146
Processing:   201  262  29.1    262     717
Waiting:      201  262  29.1    262     717
Total:        201  263  34.3    262     854

Percentage of the requests served within a certain time (ms)
  50%    262
  66%    269
  75%    275
  80%    279
  90%    292
  95%    307
  98%    332
  99%    380
 100%    854 (longest request)

---------------------------------------------------------------

Server Software:
Server Hostname:        192.168.18.213
Server Port:            32555

Document Path:          /bench200
Document Length:        18 bytes

Concurrency Level:      2000
Time taken for tests:   28.870 seconds
Complete requests:      100000
Failed requests:        0
Write errors:           0
Keep-Alive requests:    100000
Total transferred:      16500000 bytes
HTML transferred:       1800000 bytes
Requests per second:    3463.86 [#/sec] (mean)
Time per request:       577.391 [ms] (mean)
Time per request:       0.289 [ms] (mean, across all concurrent requests)
Transfer rate:          558.14 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    3  22.3      0     219
Processing:   285  511 276.7    401    6874
Waiting:      202  511 276.7    401    6874
Total:        285  514 278.2    401    6874

Percentage of the requests served within a certain time (ms)
  50%    401
  66%    416
  75%    449
  80%    613
  90%    812
  95%   1005
  98%   1361
  99%   1669
 100%   6874 (longest request)

---------------------------------------------------------------

Server Software:
Server Hostname:        192.168.18.213
Server Port:            9191

Document Path:          /bench200
Document Length:        18 bytes

Concurrency Level:      1000
Time taken for tests:   135.344 seconds
Complete requests:      523457
Failed requests:        0
Write errors:           0
Keep-Alive requests:    523457
Total transferred:      86370405 bytes
HTML transferred:       9422226 bytes
Requests per second:    3867.60 [#/sec] (mean)
Time per request:       258.558 [ms] (mean)
Time per request:       0.259 [ms] (mean, across all concurrent requests)
Transfer rate:          623.20 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0  14.0      0    1007
Processing:   201  257  34.4    257     484
Waiting:      194  257  34.4    257     484
Total:        201  258  37.2    257    1303

Percentage of the requests served within a certain time (ms)
  50%    257
  66%    272
  75%    281
  80%    287
  90%    303
  95%    318
  98%    334
  99%    345
 100%   1303 (longest request)
 
---------------------------------------------------------------

Server Software:
Server Hostname:        192.168.18.213
Server Port:            9191

Document Path:          /bench200
Document Length:        18 bytes

Concurrency Level:      2000
Time taken for tests:   288.130 seconds
Complete requests:      908503
Failed requests:        0
Write errors:           0
Keep-Alive requests:    908503
Total transferred:      149902995 bytes
HTML transferred:       16353054 bytes
Requests per second:    3153.10 [#/sec] (mean)
Time per request:       634.296 [ms] (mean)
Time per request:       0.317 [ms] (mean, across all concurrent requests)
Transfer rate:          508.07 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   8.5      0     287
Processing:   340  632 382.6    473   43318
Waiting:      206  632 382.6    473   43318
Total:        340  632 382.9    473   43318

Percentage of the requests served within a certain time (ms)
  50%    473
  66%    507
  75%    765
  80%    797
  90%   1095
  95%   1367
  98%   1788
  99%   2071
 100%  43318 (longest request)
 
---------------------------------------------------------------

Server Software:        nginx/1.13.3
Server Hostname:        192.168.18.213
Server Port:            80

Document Path:          /bench200
Document Length:        18 bytes

Concurrency Level:      1000
Time taken for tests:   176.627 seconds
Complete requests:      275211
Failed requests:        0
Write errors:           0
Keep-Alive requests:    273211
Total transferred:      69343172 bytes
HTML transferred:       4953798 bytes
Requests per second:    1558.15 [#/sec] (mean)
Time per request:       641.787 [ms] (mean)
Time per request:       0.642 [ms] (mean, across all concurrent requests)
Transfer rate:          383.40 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2  27.9      0    3124
Processing:   210  638 200.9    581   13410
Waiting:      206  638 200.9    581   13410
Total:        210  640 202.3    582   13410

Percentage of the requests served within a certain time (ms)
  50%    582
  66%    595
  75%    607
  80%    618
  90%    901
  95%    993
  98%   1294
  99%   1578
 100%  13410 (longest request)

---------------------------------------------------------------

Server Software:        nginx/1.13.3
Server Hostname:        192.168.18.213
Server Port:            80

Document Path:          /bench200
Document Length:        18 bytes

Concurrency Level:      2000
Time taken for tests:   176.885 seconds
Complete requests:      269913
Failed requests:        580
   (Connect: 0, Receive: 0, Length: 580, Exceptions: 0)
Write errors:           0
Non-2xx responses:      579
Keep-Alive requests:    267932
Total transferred:      68099406 bytes
HTML transferred:       4953951 bytes
Requests per second:    1525.92 [#/sec] (mean)
Time per request:       1310.685 [ms] (mean)
Time per request:       0.655 [ms] (mean, across all concurrent requests)
Transfer rate:          375.97 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    8 183.0      0   31220
Processing:   342 1289 1289.9    954   75000
Waiting:        0 1289 1282.1    954   62543
Total:        342 1297 1301.8    956   75000

Percentage of the requests served within a certain time (ms)
  50%    956
  66%   1308
  75%   1590
  80%   1649
  90%   2309
  95%   3033
  98%   4355
  99%   5399
 100%  75000 (longest request)

```

