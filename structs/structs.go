package structs

import "gopkg.in/guregu/null.v4"

type FIO struct {
	ID          uint   `json:"id,omitempty"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}

type EditFIO struct {
	ID          uint        `json:"id,omitempty"`
	Name        null.String `json:"name"`
	Surname     null.String `json:"surname"`
	Patronymic  null.String `json:"patronymic,omitempty"`
	Age         null.Int    `json:"age,omitempty"`
	Gender      null.String `json:"gender,omitempty"`
	Nationality null.String `json:"nationality,omitempty"`
}

type SearchUserData struct {
	Name        null.String `json:"name"`
	Surname     null.String `json:"surname"`
	Patronymic  null.String `json:"patronymic"`
	Age_down    null.Int    `json:"age_down"`
	Age_up      null.Int    `json:"age_up"`
	Gender      null.String `json:"gender"`
	Nationality null.String `json:"nationality"`
}
