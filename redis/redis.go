package redis

import (
	"context"
	"fio_service/config"
	"fio_service/structs"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisCache struct {
	rdb *redis.Client
}

var RDB RedisCache

func InitRedis() {
	RDB.rdb = redis.NewClient(&redis.Options{
		Addr:     config.Conf.APP_RDB_ADDR,
		Password: config.Conf.RDB_PASSWORD, // no password set
		DB:       0,                        // use default DB
	})
}

func (r *RedisCache) SetUser(f structs.FIO) error {
	return r.rdb.HSet(ctx, fmt.Sprintf("user:%d", f.ID),
		"name", f.Name,
		"sname", f.Surname,
		"patr", f.Patronymic,
		"gend", f.Gender,
		"age", f.Age,
		"nati", f.Nationality,
	).Err()
}

func (r *RedisCache) GetUser(id uint) (structs.FIO, error) {
	userdata := r.rdb.HGetAll(ctx, fmt.Sprintf("user:%d", id))
	if userdata.Err() != nil {
		return structs.FIO{}, userdata.Err()
	}
	age, err := strconv.Atoi(userdata.Val()["age"])
	if err != nil {
		return structs.FIO{}, err
	}
	return structs.FIO{
		ID:          id,
		Name:        userdata.Val()["name"],
		Surname:     userdata.Val()["sname"],
		Patronymic:  userdata.Val()["patr"],
		Age:         age,
		Gender:      userdata.Val()["gend"],
		Nationality: userdata.Val()["nati"],
	}, userdata.Err()
}

func (r *RedisCache) DelUser(id uint) error {
	cmd := r.rdb.Del(ctx, fmt.Sprintf("user:%d", id))
	return cmd.Err()
}

func (r *RedisCache) SetKafkaOffset(offset int64) error {
	return r.rdb.Set(ctx, "kafka", offset, 0).Err()
}

func (r *RedisCache) GetKafkaOffset() int64 {
	cmd := r.rdb.Get(ctx, "kafka")
	offset, err := strconv.ParseInt(cmd.Val(), 10, 64)
	if err != nil {
		return 0
	}
	return offset
}
