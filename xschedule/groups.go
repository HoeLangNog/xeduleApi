package xschedule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Group struct {
	Code string `json:"code"`
	Id   string `json:"id"`
	Orus []int  `json:"orus"`
}

var GroupCache []*Group
var lastPulledGroupCache *time.Time

func GetAllGroups() []*Group {
	if lastPulledGroupCache != nil && lastPulledGroupCache.Unix() > time.Now().Unix()-300 {
		return GroupCache
	} else {
		if GroupCache == nil {
			return pullGroups()
		} else {
			a := time.Now()
			lastPulledGroupCache = &a
			go func() {
				_ = pullGroups()
			}()
			return GroupCache
		}
	}

}

func GetGroup(code string) *Group {
	groups := GetAllGroups()

	for _, group := range groups {
		if group.Code == code {
			return group
		}
	}
	return nil
}

func GetGroupById(id string) *Group {
	groups := GetAllGroups()

	for _, group := range groups {
		if group.Id == id {
			return group
		}
	}
	return nil
}

func pullGroups() []*Group {
	client := GetAndCheckCookies()

	get, err := client.Get("https://sa-curio.xedule.nl/api/group/")

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if get.StatusCode != http.StatusOK {
		Login()
		return GetAllGroups()
	}

	d := json.NewDecoder(get.Body)

	var groups []*Group
	err = d.Decode(&groups)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	GroupCache = groups
	a := time.Now()
	lastPulledGroupCache = &a
	return groups
}
