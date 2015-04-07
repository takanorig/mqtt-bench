package main

import (
	"bytes"
	"flag"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"os"
	"strconv"
	"sync"
	"time"
)

const BASE_TOPIC string = "/mqtt-bench/benchmark"

var Debug bool = false

// 実行オプション
type ExecOptions struct {
	Broker       string // Broker URI
	Qos          byte   // QoS(0/1/2)
	Retain       bool   // Retain
	Topic        string // Topicのルート
	Username     string // ユーザID
	Password     string // パスワード
	ClientNum    int    // クライアントの同時実行数
	Count        int    // 1クライアント当たりのメッセージ数
	MessageSize  int    // 1メッセージのサイズ(byte)
	PreTime      int    // 実行前の待機時間(ms)
	IntervalTime int    // メッセージ毎の実行間隔時間(ms)
}

func Execute(exec func(clients []*MQTT.Client, opts ExecOptions, param ...string), opts ExecOptions) {
	message := CreateFixedSizeMessage(opts.MessageSize)

	clients := make([]*MQTT.Client, opts.ClientNum)
	hasErr := false
	for i := 0; i < opts.ClientNum; i++ {
		client := Connect(opts.Broker, opts.Username, opts.Password, i)
		if client == nil {
			hasErr = true
			break
		}
		clients[i] = client
	}

	// 接続エラーがあれば、接続済みのクライアントの切断処理を行い、処理を終了する。
	if hasErr {
		for i := 0; i < len(clients); i++ {
			client := clients[i]
			if client != nil {
				Disconnect(client)
			}
		}
		return
	}

	// 安定させるために、一定時間待機する。
	time.Sleep(time.Duration(opts.PreTime) * time.Millisecond)

	fmt.Printf("%s Start benchmark\n", time.Now())

	startTime := time.Now()
	exec(clients, opts, message)
	endTime := time.Now()

	fmt.Printf("%s End benchmark\n", time.Now())

	// 切断に時間がかかるため、非同期で処理を行う。
	AsyncDisconnect(clients)

	// 処理結果を出力する。
	totalCount := opts.ClientNum * opts.Count
	duration := (endTime.Sub(startTime)).Nanoseconds() / int64(1000000) // nanosecond -> millisecond
	throughput := float64(totalCount) / float64(duration) * 1000        // messages/sec
	fmt.Printf("\nResult : broker=%s, clients=%d, totalCount=%d, duration=%dms, throughput=%.2fmessages/sec\n",
		opts.Broker, opts.ClientNum, totalCount, duration, throughput)
}

// 全クライアントに対して、publishの処理を行う。
// 送信したメッセージ数を返す（原則、クライアント数分となる）。
func PublishAllClient(clients []*MQTT.Client, opts ExecOptions, param ...string) {
	message := param[0]

	wg := new(sync.WaitGroup)

	for id := 0; id < len(clients); id++ {
		wg.Add(1)

		client := clients[id]

		go func(clientId int) {
			defer wg.Done()

			for index := 0; index < opts.Count; index++ {
				topic := fmt.Sprintf(opts.Topic+"/%d", clientId)

				if Debug {
					fmt.Printf("Publish : id=%d, count=%d, topic=%s\n", clientId, index, topic)
				}
				Publish(client, topic, opts.Qos, opts.Retain, message)

				if opts.IntervalTime > 0 {
					time.Sleep(time.Duration(opts.IntervalTime) * time.Millisecond)
				}
			}
		}(id)
	}

	wg.Wait()
}

// メッセージを送信する。
func Publish(client *MQTT.Client, topic string, qos byte, retain bool, message string) {
	token := client.Publish(topic, qos, retain, message)

	if token.Wait() && token.Error() != nil {
		fmt.Printf("Publish error: %s\n", token.Error())
	}
}

// 全クライアントに対して、subscribeの処理を行う。
// 受信したメッセージ数を返す。
func SubscribeAllClient(clients []*MQTT.Client, opts ExecOptions, param ...string) {
	wg := new(sync.WaitGroup)

	for id := 0; id < len(clients); id++ {
		wg.Add(1)

		client := clients[id]

		go func(clientId int) {
			defer wg.Done()

			for index := 0; index < opts.Count; index++ {
				topic := fmt.Sprintf(opts.Topic+"/%d", clientId)

				if Debug {
					fmt.Printf("Subscribe : id=%d, count=%d, topic=%s\n", clientId, index, topic)
				}
				Subscribe(client, topic, opts.Qos)

				if opts.IntervalTime > 0 {
					time.Sleep(time.Duration(opts.IntervalTime) * time.Millisecond)
				}
			}
		}(id)
	}

	wg.Wait()
}

// メッセージを受信する。
func Subscribe(client *MQTT.Client, topic string, qos byte) {
	var handler MQTT.MessageHandler = func(client *MQTT.Client, msg MQTT.Message) {
		if Debug {
			fmt.Printf("Received message : topic=%s, message=%s\n", msg.Topic(), msg.Payload())
		}
	}

	token := client.Subscribe(topic, qos, handler)

	if token.Wait() && token.Error() != nil {
		fmt.Printf("Subscribe error: %s\n", token.Error())
	}
}

// 固定サイズのメッセージを生成する。
func CreateFixedSizeMessage(size int) string {
	var buffer bytes.Buffer
	for i := 0; i < size; i++ {
		buffer.WriteString(strconv.Itoa(i % 10))
	}

	message := buffer.String()
	return message
}

// 指定されたBrokerへ接続し、そのMQTTクライアントを返す。
// 接続に失敗した場合は nil を返す。
func Connect(broker string, username string, password string, id int) *MQTT.Client {

	// 複数プロセスで、ClientIDが重複すると、Broker側で問題となるため、
	// プロセスIDを利用して、IDを割り振る。
	// mqttbench<プロセスIDの16進数値>-<クライアントの連番>
	pid := strconv.FormatInt(int64(os.Getpid()), 16)
	clientId := fmt.Sprintf("mqttbench%s-%d", pid, id)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	client := MQTT.NewClient(opts)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		fmt.Printf("Connected error: %s\n", token.Error())
		return nil
	}

	return client
}

// 非同期でBrokerとの接続を切断する。
func AsyncDisconnect(clients []*MQTT.Client) {
	wg := new(sync.WaitGroup)

	for _, client := range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Disconnect(client)
		}()
	}

	wg.Wait()
}

// Brokerとの接続を切断する。
func Disconnect(client *MQTT.Client) {
	client.ForceDisconnect()
}

func main() {
	broker := flag.String("broker", "tcp://{host}:{port}", "URI of MQTT broker (required)")
	action := flag.String("action", "p|pub|publish or s|sub|subscribe", "Publish or Subscribe (required)")
	qos := flag.Int("qos", 0, "MQTT QoS(0|1|2)")
	retain := flag.Bool("retain", false, "MQTT Retain")
	topic := flag.String("topic", BASE_TOPIC, "Base topic")
	username := flag.String("broker-username", "", "Username for connecting to the MQTT broker")
	password := flag.String("broker-password", "", "Password for connecting to the MQTT broker")
	clients := flag.Int("clients", 10, "Number of clients")
	count := flag.Int("count", 100, "Number of loops per client")
	size := flag.Int("size", 1024, "Message size per publish (byte)")
	preTime := flag.Int("pretime", 3000, "Pre wait time (ms)")
	intervalTime := flag.Int("intervaltime", 0, "Interval time per message (ms)")
	debug := flag.Bool("x", false, "Debug mode")

	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	// validate "broker"
	if broker == nil || *broker == "" || *broker == "tcp://{host}:{port}" {
		fmt.Printf("Invalid argument : -broker -> %s\n", *broker)
		return
	}

	// validate "action"
	var method string = ""
	if *action == "p" || *action == "pub" || *action == "publish" {
		method = "pub"
	} else if *action == "s" || *action == "sub" || *action == "subscribe" {
		method = "sub"
	}

	if method != "pub" && method != "sub" {
		fmt.Printf("Invalid argument : -action -> %s\n", *action)
		return
	}

	execOpts := ExecOptions{}
	execOpts.Broker = *broker
	execOpts.Qos = byte(*qos)
	execOpts.Retain = *retain
	execOpts.Topic = *topic
	execOpts.Username = *username
	execOpts.Password = *password
	execOpts.ClientNum = *clients
	execOpts.Count = *count
	execOpts.MessageSize = *size
	execOpts.PreTime = *preTime
	execOpts.IntervalTime = *intervalTime

	Debug = *debug

	switch method {
	case "pub":
		Execute(PublishAllClient, execOpts)
	case "sub":
		Execute(SubscribeAllClient, execOpts)
	}
}
