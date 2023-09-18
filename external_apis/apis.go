package externalapis

import (
	"encoding/json"
	"errors"
	"fio_service/structs"
	"fmt"
	"io"
	"net/http"
)

// all
func GenUserData(u *structs.FIO) error {
	if err := GetAge(u); err != nil {
		return err
	}
	if err := GetGender(u); err != nil {
		return err
	}
	if err := GetNationality(u); err != nil {
		return err
	}
	return nil
}

// возраст
func GetAge(u *structs.FIO) error {
	// send request
	resp, err := http.DefaultClient.Get(fmt.Sprintf("https://api.agify.io/?name=%s", u.Name))
	if err != nil {
		return err
	}
	// get body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// decode json
	var a AgeApiResponse
	err = json.Unmarshal(body, &a)
	if err != nil {
		return err
	}
	u.Age = a.Age
	// return
	return err
}

// пол
func GetGender(u *structs.FIO) error {
	// send request
	resp, err := http.DefaultClient.Get(fmt.Sprintf("https://api.genderize.io/?name=%s", u.Name))
	if err != nil {
		return err
	}
	// get body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// decode json
	var g GenderApiResponse
	err = json.Unmarshal(body, &g)
	if err != nil {
		return err
	}
	u.Gender = g.Gender
	// return
	return err
}

// национальность
func GetNationality(u *structs.FIO) error {
	// send request
	resp, err := http.DefaultClient.Get(fmt.Sprintf("https://api.nationalize.io/?name=%s", u.Name))
	if err != nil {
		return err
	}
	// get body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// decode json
	var n NationalityApiResponse
	err = json.Unmarshal(body, &n)
	if err != nil {
		return err
	}
	// check length
	if len(n.Country) == 0 {
		return errors.New("country not found")
	}
	u.Nationality = n.Country[0].CountryID
	// return
	return err
}
