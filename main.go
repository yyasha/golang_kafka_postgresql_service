package main

import (
	"fio_service/config"
	"fio_service/http_api"
	"fio_service/kafka"
	"fio_service/postgres"
	"fio_service/redis"
	"fmt"
	"log"
)

func main() {
	log.Println("Server started")
	// load env config
	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	// connect to redis
	redis.InitRedis()
	// connect to db
	if err := postgres.InitDB(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", config.Conf.DB_USER, config.Conf.DB_PASSWORD, config.Conf.APP_DB_ADDR, config.Conf.DB_NAME), config.Conf.DB_MIGRATE_VERSION); err != nil {
		log.Fatal(err)
	}
	// start kafka consumer
	go kafka.ConsumeMessages()
	// start http server
	http_api.StartHttpServer()
}
