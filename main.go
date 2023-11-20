package main

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"go-mqtt-monitoring-server/logger"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const mqttServer = "127.0.0.1:1883"
const mqttClientID = "some-unique-string"

/*
вся бизнес-логика связана с этим внутренним клиентом MQTT mqttClientID. Этот клиент играет центральную роль в обработке сообщений, управлении подписками на топики и отправке сообщений на брокер MQTT. В обработчиках, таких как temperatureCallback, monitorCallback и newObjectRegistryCallback, осуществляется логика, связанная с приемом и обработкой сообщений внутри топиков.

Он также участвует в создании и управлении топиками для новых клиентов, как было добавлено в предыдущих изменениях кода.
*/

const tempTopic = "/temperature"
const actionTopic = "/action"
const monitorTopic = "/monitor"                       //Вывод данных в консоль при изменениях в temperature
const newObjectRegistryTopic = "/new_object_registry" //Регистрация новой "машинки"
const stateTopicPrefix = "/state/"                    //Каталог топиков со статусами "машинок"
const actionTopicPrefix = "/action/"                  //Каталог топиков с командами для "машинок"

var minTemp float64 = 28.0
var maxTemp float64 = 29.0

var wg = sync.WaitGroup{}
var existingClients = make(map[string]bool) //Здесь будет массив клиентов. Только его нужно превратить в объект БД
//И хранить их в БД, а не в переменной.
var mutex = &sync.Mutex{} //Объект mutex для транзакционной работы с новыми клиентами

func main() {
	wg.Add(1)

	greeter()

	c := createClient()
	//В зависимости от созданной подписки
	//Если обращение в топик tempTopic = "/temperature", вызываем коллбэк temperatureCallback
	if token := c.Subscribe(tempTopic, 0, temperatureCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	//Если обращение в топик monitorTopic = "/monitor", вызываем коллбэк monitorCallback
	if token := c.Subscribe(monitorTopic, 0, monitorCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	//Если обращение в топик newObjectRegistryTopic = "/new_object_registry", вызываем коллбэк newObjectRegistryCallback
	if token := c.Subscribe(newObjectRegistryTopic, 0, newObjectRegistryCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	wg.Wait()
}

//Создаём внутреннего клиента для реализации бизнес-логики
/*
mqtt.NewClientOptions(): Создает новый объект настроек для MQTT-клиента.
AddBroker("tcp://" + mqttServer): Устанавливает адрес брокера MQTT для подключения.
SetClientID(mqttClientID): Устанавливает идентификатор клиента MQTT.
opts.AutoReconnect = true: Включает автоматическое переподключение к брокеру в случае разрыва соединения.
Функция createClient() в вашем коде отвечает за создание и настройку нового объекта MQTT-клиента (c).*/
func createClient() mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + mqttServer).SetClientID(mqttClientID)
	opts.AutoReconnect = true

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return c
}

func temperatureCallback(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()
	actionHandler(client, string(payload))
	killSwitch(string(payload))
}

func monitorCallback(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()
	monitorHandler(string(payload))
	killSwitch(string(payload))
}

func newObjectRegistryCallback(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()
	clientID := string(payload)
	if !clientExists(clientID) {
		createClientTopics(client, clientID)
		client.Publish(newObjectRegistryTopic, 0, false, "Topics created for "+clientID)
	} else {
		fmt.Println("Client already exists: " + clientID)
		client.Publish(newObjectRegistryTopic, 0, false, "Client "+clientID+" already exists")
	}
}

func actionHandler(client mqtt.Client, payload string) {
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

func monitorHandler(payload string) {
	if strings.Compare(payload, "\n") > 0 {
		t := time.Now()
		data := "[" + t.Format("2006-01-02 15:04:05") + "] monitor: " + payload
		fmt.Println("["+t.Format("2006-01-02 15:04:05")+"]", "monitor: ", payload)
		tg := logger.TelegramLogger{}
		tg.Init().Log(data)
	}
}

func killSwitch(payload string) {
	if strings.Compare("bye", string(payload)) == 0 {
		fmt.Println("exiting . . .")
		wg.Done()
	}
}

func greeter() {
	fmt.Println("==============================================")
	fmt.Println("* * * Привет от сервиса мониторинга MQTT * * *")
	fmt.Println("======      сервис успешно запущен      ======")
	fmt.Println("==============================================")
}

func createClientTopics(client mqtt.Client, clientID string) {
	stateTopic := stateTopicPrefix + clientID
	actionTopic := actionTopicPrefix + clientID

	mutex.Lock()
	existingClients[clientID] = true
	mutex.Unlock()

	if token := client.Subscribe(stateTopic, 0, stateCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := client.Subscribe(actionTopic, 0, temperatureCallback); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func stateCallback(client mqtt.Client, msg mqtt.Message) {
	// Handle state messages if needed
}

func clientExists(clientID string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	_, exists := existingClients[clientID]
	return exists
}
