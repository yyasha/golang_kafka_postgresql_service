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

// func ExampleClient() {
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})

// 	err := rdb.Set(ctx, "key", "value", 0).Err()
// 	if err != nil {
// 		panic(err)
// 	}

// 	val, err := rdb.Get(ctx, "key").Result()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("key", val)

// 	val2, err := rdb.Get(ctx, "key2").Result()
// 	if err == redis.Nil {
// 		fmt.Println("key2 does not exist")
// 	} else if err != nil {
// 		panic(err)
// 	} else {
// 		fmt.Println("key2", val2)
// 	}
// }
