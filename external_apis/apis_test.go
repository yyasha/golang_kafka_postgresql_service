package externalapis

import (
	"fio_service/structs"
	"testing"
)

var tests = []structs.FIO{
	{Name: "Dmitriy"},
	{Name: "Mikael"},
	{Name: "Maxim"},
}

func TestApis(t *testing.T) {
	for _, test := range tests {
		err := GenUserData(&test)
		if err != nil {
			t.Error(err)
		}
		if test.Age == 0 {
			t.Error("age not generated")
		}
		if test.Gender == "" {
			t.Error("gender not generated")
		}
		if test.Nationality == "" {
			t.Error("nationality not generated")
		}
	}
}
