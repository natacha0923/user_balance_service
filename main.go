package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"

	"github.com/natacha0923/user_balance_service/manager"
	"github.com/natacha0923/user_balance_service/migrations"
	"github.com/natacha0923/user_balance_service/pump"
)

func main() {
	err := app()
	if err != nil {
		log.Fatal(err)
	}
}

func app() error {
	log.Println("waiting for rabbit at least 15 seconds")
	time.Sleep(15 * time.Second)
	pgConnStr, ok := os.LookupEnv("POSTGRES_CONN_STR")
	if !ok {
		return fmt.Errorf("failed to get a postgres connection string from env")
	}
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	err = migrations.Run(db)
	if err != nil {
		return fmt.Errorf("failed to apply migration to postgres: %v", err)
	}

	amqpConnStr, ok := os.LookupEnv("AMQP_CONN_STR")
	if !ok {
		return fmt.Errorf("failed to get a amqp connection string from env")
	}
	conn, err := amqp.Dial(amqpConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	man := &manager.UserBalanceManager{
		Db: db,
	}

	msgPump := pump.MessagePump{
		AMQPConn: conn,
		Manager:  man,
	}
	return msgPump.Run()
}
