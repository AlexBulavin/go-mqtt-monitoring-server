# Чтобы запустить сервис на локальной машине:

Создаём глобальную папку проекта.
Например, mqtt_services:
>mkdir mqtt_services

Перемещаемся в неё:
>cd mqtt_services

Клонируем из github настоящий репозиторий (предполагается, что сам git на локальной машине уже развёрнут):
>git clone https://github.com/AlexBulavin/go-mqtt-monitoring-server.git


# Go MQTT monitoring server

Это элементарный основанный на MQTT golang сервис backend для мониторинга, управления и логирования датчиков температуры.

Он получает постоянный поток публикации данных в топике и выполняет actions основанные на желаемой логике, реализованном на бэкенде.
После этого он логирует подтверждение в онлайн лог телеграм бота с помощью Bot API

It gets the continuously streamed/published data from an MQTT topic and then takes action based on desired logic implemented on the back-end.
After that it will log the acknowledge to an online log channel based on a Telegram Bot, with the help of the official Bot API. 

Для корректной работы сервиса мониторинга необходимо MQTT брокекр.
Устанавливаем его на imac так (через homebrew):
>brew install mosquitto

Узнаём его папку расположения:
>find /usr/local/Cellar/mosquitto -name mosquitto

Получим ответ в виде:
/usr/local/Cellar/mosquitto
/usr/local/Cellar/mosquitto/2.0.18/sbin/mosquitto
/usr/local/Cellar/mosquitto/2.0.18/etc/mosquitto

Запускаем брокер:
>/usr/local/Cellar/mosquitto/2.0.18/sbin/mosquitto

Если нужно задать хост и порт в явной форме, это делается так:
> /usr/local/Cellar/mosquitto/2.0.18/sbin/mosquitto -p <PORT> -h <HOST>

Например:
>/usr/local/Cellar/mosquitto/2.0.18/sbin/mosquitto -p 1883 -h 192.168.110.2

Если вы хотите, чтобы брокер слушал соединения на всех доступных интерфейсах, используйте 0.0.0.0.
Пример, если вы хотите запустить Mosquitto на порту 1883 и слушать соединения на всех интерфейсах:
>/usr/local/Cellar/mosquitto/2.0.18/sbin/mosquitto -p 1883 -h 0.0.0.0

Чтобы остановить брокера используем команду:
>pkill mosquitto

или
>brew services stop mosquitto 

Далее при запущенном MQTT брокере запускаем сервис мониторинга настоящего проекта, прописывая его 
хост и порт таким, какой задали для брокера.
По умолчанию (если ничего не задавать на брокере) это 127.0.0.1:1883

# Associated resources
- Introduction and review video: [youtu.be/zXzmXzBmWdY](https://youtu.be/zXzmXzBmWdY)
- ESP8266 code repo: [AyubIRZ/esp8266-mqtt-temperature-monitoring](https://github.com/AyubIRZ/esp8266-mqtt-temperature-monitoring)

## Notes
- This is a simple demonstration of the project. You may not use it in production without reviewing the code and changing it to the proper version of your use!
- Data transmission between clients is not secured with TLS or authentication! It's simple and dumb!
-  Any contribution, PR or submitting issues using github issue tracker will be appreciated!
