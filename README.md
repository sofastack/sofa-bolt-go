# sofa-bolt-go
The Golang implementation of the SOFABolt protocol.
=======
# TOC

-   [synopsis](#synopsis)
-   [example](#example)
    -   [command](#command)
    -   [client & server](#client---server)
-   [cli](#cli)
    -   [install](#install)
    -   [decode](#decode)
-   [benchmark](#benchmark)

# synopsis

sofa-bolt-go 是 [bolt 1.0/2.0 serialization protocol](https://www.sofastack.tech/projects/sofa-bolt/overview/) 以及 TBRemoting 的 Golang 实现，包含了客户端/服务端/编解码。

sofa-bolt-go 提供三种类型的 API:

-   command: 编解码 bolt 1.0/2.0 以及 TaobaoRemoting 协议。
-   client: 发送请求并接受响应，同时也支持接受请求并写会响应。
-   server: 解析请求并写会响应.

# example

## client & server

See [client_server_example_test.go](/examples/client_server_example_test.go#L12)

## client & server with dialer

See [client_server_example_test.go](/examples/client_server_example_test.go#L67)

## command

See [command_example_test.go](/sofabolt/command_example_test.go#L9)

# cli

## install

`make build` 或者 `go get github.com/sofastack/sofa-bolt-go/cmd/bolt`

## decode

```bash
bin/bolt decode 0100000200000000020000c80000000a0000000b0000000131000000013168656c6c6f20776f726c64
```

## decodeheader

```bash
bin/bolt decodeheader 00000018736f66615f686561645f7461726765745f736572766963650000001048656c6c6f536572766963653a312e300000001b7270635f74726163655f636f6e746578742e736f666152706349640000000130000000167270635f74726163655f636f6e746578742e73616d700000000566616c73650000001d7270635f74726163655f636f6e746578742e736f6661547261636549640000001e6139666531363839313537313239333033373932323130313233383031330000001f7270635f74726163655f636f6e746578742e736f666143616c6c6572496463000000000000001e7270635f74726163655f636f6e746578742e736f666143616c6c65724970000000000000001e7270635f74726163655f636f6e746578742e736f666150656e417474727300000000000000207270635f74726163655f636f6e746578742e736f666143616c6c65725a6f6e650000000000000014736f66615f686561645f7461726765745f617070000000000000000870726f746f636f6c00000004626f6c7400000007736572766963650000001048656c6c6f536572766963653a312e300000001d7270635f74726163655f636f6e746578742e73797350656e4174747273000000000000001f7270635f74726163655f636f6e746578742e736f666143616c6c65724170700000000000000015736f66615f686561645f6d6574686f645f6e616d650000000873617948656c6c6f
```

# benchmark

```bash
pkg: github.com/sofastack/sofa-bolt-go/sofabolt
BenchmarkClientConcurrent-8        	  522289	      1996 ns/op	       0 B/op	       0 allocs/op
BenchmarkClient-8                  	  281299	      3987 ns/op	       0 B/op	       0 allocs/op
BenchmarkReadTBRemotingRequest/request-8         	  431926	      2385 ns/op	     384 B/op	      19 allocs/op
BenchmarkReadTBRemotingRequest/command-8         	 8268426	       145 ns/op	       0 B/op	       0 allocs/op
BenchmarkWriteRequest-8                          	29614092	        39.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkWriteResponse-8                         	17844903	        64.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkServer/1_connection-8                   	 2368573	       505 ns/op	       0 B/op	       0 allocs/op
BenchmarkServer/128_connection-8                 	 1862937	       786 ns/op	       0 B/op	       0 allocs/op
BenchmarkServer/512_connection-8                 	 3341610	       343 ns/op	       0 B/op	       0 allocs/op
BenchmarkServer/1024_connection-8                	 5492541	       212 ns/op	       0 B/op	       0 allocs/op
```
