package xschedule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var teacherNames []*TeacherNameEntry

type Teacher struct {
	Code string `json:"code"`
	Id   string `json:"id"`
}

type TeacherNameEntry struct {
	Code      string `json:"code"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var TeacherCache []*Teacher
var lastPulledTeacherCache *time.Time

func GetAllTeachers() []*Teacher {
	if lastPulledTeacherCache != nil && lastPulledTeacherCache.Unix() > time.Now().Unix()-300 {
		return TeacherCache
	} else {
		if TeacherCache == nil {
			return pullTeachers()
		} else {
			a := time.Now()
			lastPulledTeacherCache = &a
			go func() {
				_ = pullTeachers()
			}()
			return TeacherCache
		}
	}
}

func GetTeacherById(id string) *Teacher {
	for _, t := range GetAllTeachers() {
		if t.Id == id {
			return t
		}
	}
	return nil
}

func pullTeachers() []*Teacher {
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

func GetTeacherName(code string) (firstname string, lastname string, found bool) {
	for _, nameEntry := range teacherNames {
		if nameEntry.Code == code {
			return nameEntry.FirstName, nameEntry.LastName, true
		}
	}
	return "", "", false
}
