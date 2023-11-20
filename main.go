package main

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"go-mqtt-monitoring-server/logger" // change this path to your local GOPATH format. ex: myProjectDir/logger
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MQTT credentials(you may have username and password too)
const mqttServer = "127.0.0.1:1883"       //Эта константа определяет адрес (IP и порт) брокера MQTT, к которому программа будет подключаться. В данном случае, брокер находится по IP-адресу 192.168.100.2 и слушает порт 1883.
const mqttClientID = "some-unique-string" //Идентификатор клиента MQTT:

// MQTT topics(channels) that we work with.
const tempTopic = "/temperature"
const actionTopic = "/action"
const monitorTopic = "/monitor"

// temperature thresholds that we take actions based on.
var minTemp float64 = 28.0
var maxTemp float64 = 29.0

var wg = sync.WaitGroup{}

func main() {
	wg.Add(1)

	greeter()

	c := createClient()

	if token := c.Subscribe(tempTopic, 0, actionCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := c.Subscribe(monitorTopic, 0, monitorCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	wg.Wait()
}

// createClient returns a new MQTT client object.
func createClient() MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker("tcp://" + mqttServer).SetClientID(mqttClientID)
	opts.AutoReconnect = true

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return c
}

// define a function for the Action message handler
func actionCallback(client MQTT.Client, msg MQTT.Message) {
	payload := msg.Payload()
	actionHandler(client, string(payload))
	killSwitch(string(payload))
}

// define a function for the Monitor message handler
func monitorCallback(client MQTT.Client, msg MQTT.Message) {
	payload := msg.Payload()
	monitorHandler(string(payload))
	killSwitch(string(payload))
}

// actionHandler defines and executes the logic for each incoming message on /action topic
func actionHandler(client MQTT.Client, payload string) {
	temperature, err := strconv.ParseFloat(payload, 64)
	if err != nil {
		panic(err.Error())
	}

	if strings.Compare(payload, "\n") > 0 {

		t := time.Now()
		fmt.Println("["+t.Format("2006-01-02 15:04:05")+"]", "temperature: ", payload)

		switch {
		case temperature <= minTemp:
			client.Publish(actionTopic, 0, false, "-1")

		case temperature > minTemp && temperature < maxTemp:
			client.Publish(actionTopic, 0, false, "0")

		case temperature >= maxTemp:
			client.Publish(actionTopic, 0, false, "1")
		}
	}
}

// monitorHandler defines and executes the logic for each incoming message on /monitor topic
func monitorHandler(payload string) {
	if strings.Compare(payload, "\n") > 0 {
		t := time.Now()
		data := "[" + t.Format("2006-01-02 15:04:05") + "] monitor: " + payload
		fmt.Println("["+t.Format("2006-01-02 15:04:05")+"]", "monitor: ", payload)
		tg := logger.TelegramLogger{}
		tg.Init().Log(data)
	}
}

// killSwitch checks for the Bye command to close the MQTT connection and the app
func killSwitch(payload string) {
	if strings.Compare("bye", string(payload)) == 0 {
		fmt.Println("exiting . . .")
		wg.Done()
	}
}

// greeter prints a short introduction text to the terminal.
func greeter() {
	fmt.Println("=============================================")
	fmt.Println("* * * HELLO FROM MQTT MONITORING SERVER * * *")
	fmt.Println("=============================================")
}

/*
Используйте команду netstat в командной строке, чтобы посмотреть, какие порты открыты на вашей локальной машине. Например:
netstat -tuln

Используйте команды для просмотра процессов и поиска брокера MQTT. Например:

ps aux | grep mosquitto

Вы можете использовать MQTT-клиент, такой как Mosquitto Sub или другой, чтобы подключиться к брокеру на вашей локальной машине и проверить, отвечает ли он на запросы. Например:

mosquitto_sub -h localhost -t "test/topic"
192.168.1.110
10.13.13.51

mosquitto_sub -h 192.168.1.110:1883 -t "test/topic"

mosquitto_sub -h localhost -t "test/topic" //Подписались на test/topic

mosquitto_pub -h localhost -t "test/topic" -m "Hello, MQTT!"//Опубликовали сообщение в test/topic

Останавливаем брокера:
brew services stop mosquitto

или
pkill mosquitto


*/
