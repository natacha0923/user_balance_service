package pump

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"

	"github.com/natacha0923/user_balance_service/structs"
)

type MessagePump struct {
	AMQPConn *amqp.Connection
	Manager  Manager
}

func (mp MessagePump) Run() error {
	ch, err := mp.AMQPConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	qbc, err := ch.QueueDeclare(
		"balance.change",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	qbt, err := ch.QueueDeclare(
		"balance.transfer",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	_, err = ch.QueueDeclare(
		"balance.event",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgsbc, err := ch.Consume(
		qbc.Name, // queue
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgsbt, err := ch.Consume(
		qbt.Name, // queue
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		if msgsbt == nil && msgsbc == nil {
			break
		}

		select {
		case msg, ok := <-msgsbc:
			if !ok {
				msgsbc = nil
				continue
			}
			err := mp.changeBalance(msg, ch)
			if err != nil {
				log.Println("failed to process change balance:", string(msg.Body), " err:", err)
			}
		case msg, ok := <-msgsbt:
			if !ok {
				msgsbt = nil
				continue
			}
			err := mp.transfer(msg, ch)
			if err != nil {
				log.Println("failed to process transfer:", string(msg.Body), " err:", err)
			}
		}
	}

	return nil
}

func (mp MessagePump) changeBalance(d amqp.Delivery, ch *amqp.Channel) error {
	var rq structs.ChangeBalanceRequest
	err := json.Unmarshal(d.Body, &rq)
	if err != nil {
		return err
	}

	err = mp.Manager.ChangeBalance(rq)
	if err != nil {
		respErr := sendResponse(ch, rq.Token, d.RoutingKey, "err : "+err.Error())
		if respErr != nil {
			return respErr
		}
		return err
	}

	return sendResponse(ch, rq.Token, d.RoutingKey, "success")
}

func (mp MessagePump) transfer(d amqp.Delivery, ch *amqp.Channel) error {
	var rq structs.TransferRequest
	err := json.Unmarshal(d.Body, &rq)
	if err != nil {
		return err
	}

	err = mp.Manager.Transfer(rq)
	if err != nil {
		respErr := sendResponse(ch, rq.Token, d.RoutingKey, "err : "+err.Error())
		if respErr != nil {
			return respErr
		}
		return err
	}

	return sendResponse(ch, rq.Token, d.RoutingKey, "success")
}

func sendResponse(ch *amqp.Channel, token, routingKey, response string) error {
	resp, err := json.Marshal(&event{
		Token:      token,
		RoutingKey: routingKey,
		Status:     response,
	})
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		"balance.event",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/json",
			Body:        resp,
		},
	)
}

type Manager interface {
	ChangeBalance(request structs.ChangeBalanceRequest) error
	Transfer(request structs.TransferRequest) error
}

type event struct {
	Token      string `json:"token"`
	RoutingKey string `json:"routingKey"`
	Status     string `json:"status"`
}
