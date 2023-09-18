package http_api

import (
	"encoding/json"
	externalapis "fio_service/external_apis"
	"fio_service/postgres"
	"fio_service/redis"
	"fio_service/structs"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

func getUsers(c *fiber.Ctx) error {
	// get params
	limit := c.QueryInt("limit", 15)
	offset := c.QueryInt("offset", 0)
	var searchData structs.SearchUserData
	err := json.Unmarshal(c.Body(), &searchData)
	if err != nil {
		error_logging(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Json decode error",
		})
	}
	// get users from db
	user_list, err := postgres.DB.GetUsers(offset, limit, searchData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Database error",
		})
	}
	// add users to cache
	for _, u := range user_list {
		err := redis.RDB.SetUser(u)
		if err != nil {
			error_logging(err)
		}
	}
	// get user from cache
	// fmt.Println(redis.RDB.GetUser(1))
	// return ok
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":     false,
		"user_list": user_list,
	})
}

func addUser(c *fiber.Ctx) error {
	// get data from body
	var f structs.FIO
	err := json.Unmarshal(c.Body(), &f)
	if err != nil {
		error_logging(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Json decode error",
		})
	}
	// validate
	if f.Name == "" || f.Surname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Missing fields",
		})
	}
	// gen data
	err = externalapis.GenUserData(&f)
	if err != nil {
		error_logging(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Error with generate data",
		})
	}
	// insert into db
	err = postgres.DB.AddUser(f)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Database error",
		})
	}
	// return ok
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
	})
}

func delUser(c *fiber.Ctx) error {
	// get user id
	user_id := c.QueryInt("id", 0)
	if user_id == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Missing fields",
		})
	}
	// delete from db
	err := postgres.DB.DelUser(user_id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Database error",
		})
	}
	// delete from cache
	err = redis.RDB.DelUser(uint(user_id))
	error_logging(err)
	// return ok
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
	})
}

func editUser(c *fiber.Ctx) error {
	// get new data
	var efio structs.EditFIO
	err := json.Unmarshal(c.Body(), &efio)
	if err != nil {
		error_logging(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Json decode error",
		})
	}
	// update db
	err = postgres.DB.EditUser(efio)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   "Database error",
		})
	}
	// delete from cache
	err = redis.RDB.DelUser(efio.ID)
	error_logging(err)
	// return ok
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
	})
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
