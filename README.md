# MQTT-Bench : MQTT Benchmark Tool
This is the benchmark tool for MQTT Broker implemented by [golang](https://golang.org/).
This can benchmark the throughput for publishing and subscribing.

Supported benchmark pattern is:
* Parallel publish from clients
* Parallel subscribe from clients

## Getting started
### Installed

### Publish
* Precondition
 * The MQTT Broker is started.
```
$ mqtt-bench -broker=tcp://192.168.1.100:1883 -action=pub
2015-04-04 12:47:38.690896 +0900 JST Start benchmark
2015-04-04 12:47:38.765896 +0900 JST End benchmark

Result : broker=tcp://192.168.56.101:1883, clients=10, totalCount=1000, duration=72ms, throughput=13888.89messages/sec
```

### Subscribe
* Precondition
 * The MQTT Broker is started.
 * The MQTT broker will keep the messages. It will be published the message with retained.
```
(Keep the messages before subscribing)
$ mqtt-bench -broker=tcp://192.168.1.100:1883 -action=pub -retain=true

$ mqtt-bench -broker=tcp://{host}:{port} -action=sub
2015-04-04 12:50:27.188396 +0900 JST Start benchmark
2015-04-04 12:50:27.477896 +0900 JST End benchmark

Result : broker=tcp://192.168.1.100:1883, clients=10, totalCount=1000, duration=287ms, throughput=3484.32messages/sec
```

## Usage
```
Usage of mqtt-bench
  -action="p/pub/publish or s/sub/subscribe"  : Publish or Subscribe (required)
  -broker="tcp://{host}:{port}"               : URI of MQTT broker (required)
  -broker-password=""                         : Password for connecting to the MQTT broker
  -broker-username=""                         : Username for connecting to the MQTT broker
  -qos=0                                      : MQTT QoS(0/1/2)
  -retain=false                               : MQTT Retain
  -topic="/mqtt-bench/benchmark"              : Base topic
  -clients=10                                 : Number of clients
  -count=100                                  : Number of loops per client
  -size=1024                                  : Message size per publish (byte)
  -pretime=3000                               : Pre wait time (ms)
  -intervaltime=0                             : Interval time per message (ms)
  -x=false                                    : Debug mode
```
