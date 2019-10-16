# User Balance Service

Сервис баланса пользователей:
- списание/зачисление средств
- перевод от пользователя к пользователю

### Installation
```docker-compose up --build```

Вся конфигурация осуществляется через docker-compose.\
При старте происходит накат миграций и тестовых данных через migrations/migrations.go


### System requirements
* Docker >= 18.06
* Go >= 1.13
* Postgres >= 9.5
* Rabbit >= 3

### Examples
Взаимодействие системы осуществляется с помощью RabbitMQ. 
После проведения любой операции генерируется событие-ответ в очередь "balance.event".

#### Изменение баланса:
routing key: balance.change
```
request: {"token":"111", "body": {"userId": 1, "amount": -10}}
response: {"token":"111","routingKey":"balance.change","status":"success"}
```
```
request: {"token":"112", "body": {"userId": 1, "amount": -1000}}
response: {"token":"112","routingKey":"balance.change","status":"err : amount is bigger than balance"}
```

#### Перевод:
routing key: balance.transfer
```
request: {"token":"113", "body": {"from": 1, "to": 2, "amount": 10}
response: {"token":"113","routingKey":"balance.transfer","status":"success"}
```
```
request: {"token":"114", "body": {"from": 1, "to": 2, "amount": 10}}
response: {"token":"114","routingKey":"balance.transfer","status":"err : can not found user with id = 6"}
```
