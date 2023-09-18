package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	externalapis "fio_service/external_apis"
	"fio_service/graph/generated"
	"fio_service/graph/model"
	"fio_service/postgres"
	"fio_service/redis"
	"fio_service/structs"
	"log"
	"runtime"
	"strconv"

	"gopkg.in/guregu/null.v4"
)

func (r *fIOResolver) ID(ctx context.Context, obj *structs.FIO) (*int, error) {
	id := int(obj.ID)
	return &id, nil
}

func (r *queryResolver) GetUsers(ctx context.Context, limit *int, offset *int, searchData *model.SearchData) (*model.UsersResponse, error) {
	var sd structs.SearchUserData
	if searchData != nil {
		age_down := int64(*searchData.AgeDown)
		age_up := int64(*searchData.AgeUp)

		sd = structs.SearchUserData{
			Name:        null.StringFromPtr(searchData.Name),
			Surname:     null.StringFromPtr(searchData.Surname),
			Patronymic:  null.StringFromPtr(searchData.Patronymic),
			Gender:      null.StringFromPtr(searchData.Gender),
			Nationality: null.StringFromPtr(searchData.Nationality),
			Age_up:      null.IntFrom(age_up),
			Age_down:    null.IntFrom(age_down),
		}
	}
	// get users from db
	user_list, err := postgres.DB.GetUsers(*offset, *limit, sd)
	if err != nil {
		var msg string = "Database error"
		return &model.UsersResponse{Error: true, Msg: &msg}, nil
	}
	// add users to cache
	for _, u := range user_list {
		err := redis.RDB.SetUser(u)
		if err != nil {
			error_logging(err)
		}
	}
	//
	res := make([]*structs.FIO, len(user_list))
	for i := 0; i < len(user_list); i++ {
		res[i] = &user_list[i]
	}
	// return ok
	return &model.UsersResponse{Error: false, UserList: res}, nil
}

func (r *queryResolver) DelUser(ctx context.Context, id string) (*model.Status, error) {
	// convert
	user_id, err := strconv.Atoi(id)
	if err != nil || user_id == 0 {
		var msg string = "Missing fields"
		return &model.Status{Error: true, Msg: &msg}, err
	}
	// delete from db
	err = postgres.DB.DelUser(user_id)
	if err != nil {
		var msg string = "Database error"
		return &model.Status{Error: true, Msg: &msg}, err
	}
	// return ok
	return &model.Status{Error: false}, nil
}

func (r *queryResolver) EditUser(ctx context.Context, fio model.UpdateFio) (*model.Status, error) {
	// get new data
	var age int64
	if fio.Age != nil {
		age = int64(*fio.Age)
	}
	var efio structs.EditFIO = structs.EditFIO{
		ID:          uint(fio.ID),
		Name:        null.StringFromPtr(fio.Name),
		Surname:     null.StringFromPtr(fio.Surname),
		Patronymic:  null.StringFromPtr(fio.Patronymic),
		Age:         null.IntFrom(age),
		Gender:      null.StringFromPtr(fio.Gender),
		Nationality: null.StringFromPtr(fio.Nationality),
	}
	// update db
	err := postgres.DB.EditUser(efio)
	if err != nil {
		msg := "Database error"
		return &model.Status{Error: true, Msg: &msg}, err
	}
	// delete from cache
	err = redis.RDB.DelUser(efio.ID)
	error_logging(err)
	// return ok
	return &model.Status{Error: false}, nil
}

func (r *queryResolver) AddUser(ctx context.Context, input model.NewFio) (*model.Status, error) {
	// panic(fmt.Errorf("not implemented"))
	// get data from body
	var f structs.FIO = structs.FIO{
		Name:    input.Name,
		Surname: input.Surname,
	}
	if input.Patronymic != nil {
		f.Patronymic = *input.Patronymic
	}
	// validate
	if f.Name == "" || f.Surname == "" {
		msg := "Missing fields"
		return &model.Status{Error: true, Msg: &msg}, nil
	}
	// gen data
	err := externalapis.GenUserData(&f)
	if err != nil {
		error_logging(err)
		msg := "Error with generate data"
		return &model.Status{Error: true, Msg: &msg}, err
	}
	// insert into db
	err = postgres.DB.AddUser(f)
	if err != nil {
		msg := "Database error"
		return &model.Status{Error: true, Msg: &msg}, err
	}
	// return ok
	// return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 	"error": false,
	// })
	return &model.Status{Error: false}, nil
}

// FIO returns generated.FIOResolver implementation.
func (r *Resolver) FIO() generated.FIOResolver { return &fIOResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type fIOResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

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
