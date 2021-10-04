package xschedule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Teacher struct {
	Code string `json:"code"`
	Id   string `json:"id"`
}

var TeacherCache []*Teacher
var lastPulledTeacherCache *time.Time

func GetAllTeachers() []*Teacher {
	if lastPulledTeacherCache != nil && lastPulledTeacherCache.Unix() > time.Now().Unix()-300 {
		return TeacherCache
	}
	client := GetAndCheckCookies()

	get, err := client.Get("https://sa-curio.xedule.nl/api/docent/")

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if get.StatusCode != http.StatusOK {
		Login()
		return GetAllTeachers()
	}

	d := json.NewDecoder(get.Body)

	var teachers []*Teacher
	err = d.Decode(&teachers)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	TeacherCache = teachers
	a := time.Now()
	lastPulledTeacherCache = &a

	return teachers
}

func GetTeacherById(id string) *Teacher {
	for _, t := range GetAllTeachers() {
		if t.Id == id {
			return t
		}
	}
	return nil
}
