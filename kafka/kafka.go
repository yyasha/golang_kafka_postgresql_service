package kafka

import (
	"context"
	"encoding/json"
	"fio_service/config"
	externalapis "fio_service/external_apis"
	"fio_service/postgres"
	"fio_service/redis"
	"fio_service/structs"
	"log"
	"runtime"
	"time"

	"github.com/segmentio/kafka-go"
)

// Read messages from kafka
func ConsumeMessages() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{config.Conf.KAFKA_ADDR},
		Topic:     "FIO",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	})
	r.SetOffset(redis.RDB.GetKafkaOffset())
	log.Println("Kafka listener started")
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			error_logging(err)
			continue
		}
		log.Printf("message at offset %d: = %s\n", m.Offset, string(m.Value))
		err = redis.RDB.SetKafkaOffset(m.Offset + 1)
		error_logging(err)
		// decode json
		var u structs.FIO
		err = json.Unmarshal(m.Value, &u)
		if err != nil {
			error_logging(err)
			SendFIOError("Некорректный json")
			continue
		}
		// validate data
		if u.Name == "" || u.Surname == "" {
			SendFIOError("Нет обязательного поля")
			continue
		}
		// gen user data
		err = externalapis.GenUserData(&u)
		if err != nil {
			error_logging(err)
			SendFIOError("Ошибка с генерацией дополнительных данных")
			continue
		}
		// insert to db
		err = postgres.DB.AddUser(u)
		if err != nil {
			error_logging(err)
			SendFIOError("Ошибка при добавлении пользователя в базу данных")
			continue
		}
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

// send fio error to kafka
func SendFIOError(msg string) error {
	// connect to kafka
	conn, err := kafka.DialLeader(context.Background(), "tcp", config.Conf.KAFKA_ADDR, "FIO_FAILED", 0)
	if err != nil {
		log.Println("failed to dial leader:", err)
		return err
	}
	// send error
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.Write([]byte(msg))
	if err != nil {
		log.Println("failed to write messages:", err)
		return err
	}
	// close connection
	if err := conn.Close(); err != nil {
		log.Println("failed to close writer:", err)
		return err
	}
	return err
}

// Log errors
func error_logging(err error) {
	if err != nil {
		pc := make([]uintptr, 10)
		n := runtime.Callers(2, pc)
		frames := runtime.CallersFrames(pc[:n])
		frame, _ := frames.Next()
		// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		log.Printf("[Postgres] error on %s: %s", frame.Function, err)
	}
}
