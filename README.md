# MQTT-Bench : MQTT Benchmark Tool
This is the benchmark tool for MQTT Broker implemented by [golang](https://golang.org/).
This can benchmark the throughput for publishing and subscribing.

Supported benchmark pattern is:
* Parallel publish from clients
* Parallel subscribe from clients with publishing

## Getting started
### Installation

Use ```go get``` and ```go install```

```
$ go get github.com/takanorig/mqtt-bench
$ go install github.com/takanorig/mqtt-bench
```

or 

Download here

https://github.com/takanorig/mqtt-bench/wiki/Download

### Publish
* Precondition
 * The MQTT Broker is started.
```
$ mqtt-bench -broker=tcp://192.168.1.100:1883 -action=pub
2015-04-04 12:47:38.690896 +0900 JST Start benchmark
2015-04-04 12:47:38.765896 +0900 JST End benchmark

Result : broker=tcp://192.168.1.100:1883, clients=10, totalCount=1000, duration=72ms, throughput=13888.89messages/sec
```

### Subscribe
* Precondition
 * The MQTT Broker is started.
 * Publishing to MQTT Broker.
```
(Publish the messages while subscribing)
$ mqtt-bench -broker=tcp://192.168.1.100:1883 -action=pub -count=10000

$ mqtt-bench -broker=tcp://192.168.1.100:1883 -action=sub
2015-04-04 12:50:27.188396 +0900 JST Start benchmark
2015-04-04 12:50:27.477896 +0900 JST End benchmark

Result : broker=tcp://192.168.1.100:1883, clients=10, totalCount=1000, duration=287ms, throughput=3484.32messages/sec
```

If the following message is output to the console, the count is over limit.
So, please set ```-intervalTime``` option. 
```
panic: Subscribe error : Not finished in the max count. It may not be received the message.
```

### TLS mode
Use ```-tls``` option.

- Server authentication
```
-tls=server:certFile
```

- Server and Client authentication
```
-tls=client:rootCAFile,clientCertFile,clientKeyFile
```

## Usage
```
Usage of mqtt-bench
  -action="p|pub or s|sub"                    : Publish or Subscribe (required)
  -broker="tcp://{host}:{port}"               : URI of MQTT broker (required)
  -broker-password=""                         : Password for connecting to the MQTT broker
  -broker-username=""                         : Username for connecting to the MQTT broker
  -tls=""                                     : TLS mode. 'server:certFile' or 'client:rootCAFile,clientCertFile,clientKeyFile'
  -qos=0                                      : MQTT QoS(0|1|2)
  -retain=false                               : MQTT Retain
  -topic="/mqtt-bench/benchmark"              : Base topic
  -clients=10                                 : Number of clients
  -count=100                                  : Number of loops per client
  -size=1024                                  : Message size per publish (byte)
  -pretime=3000                               : Pre wait time (ms)
  -intervaltime=0                             : Interval time per message (ms)
  -x=false                                    : Debug mode
```

## Note
* Using Apollo
 * If you use [Apollo 1.7.x](http://activemq.apache.org/apollo/), the subscribed messages can't be output to console even if debug mode. If you want to output the subscribed messages, designate ```-support-unknown-received``` option.
