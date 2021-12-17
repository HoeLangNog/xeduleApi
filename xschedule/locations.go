package xschedule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Location struct {
	Code string `json:"code"`
	Id   string `json:"id"`
	Orus []int  `json:"orus"`
}

var Cache []*Location
var lastPulled *time.Time

func GetAllLocations() []*Location {
	if lastPulled != nil && lastPulled.Unix() > time.Now().Unix()-300 {
		return Cache
	} else {
		if Cache == nil {
			return pullAllLocations()
		} else {
			lastPulled1 := time.Now()
			lastPulled = &lastPulled1
			go func() {
				_ = pullAllLocations()
			}()
			return Cache
		}
	}

}

func GetAllLocationsWithPrefix(prefix string, ignoreOLC bool) []*Location {
	var newLocations []*Location
	for _, f := range GetAllLocations() {
		if strings.HasPrefix(f.Code, prefix) {
			found15 := false
			for _, c := range f.Orus {
				if c == 15 {
					found15 = true
				}
			}
			if !found15 {
				continue
			}
			if strings.HasSuffix(f.Code, "examen") {
				continue
			}
			if ignoreOLC && strings.HasPrefix(f.Code, prefix+"OLC") {
				continue
			}
			newLocations = append(newLocations, f)
		}
	}
	return newLocations
}

func GetAvailableLocations(prefix string, sTime time.Time, eTime time.Time) []*Location {
	allLocs := GetAllLocationsWithPrefix(prefix, true)

	year, week := sTime.ISOWeek()
	var selectors []*TimeSelector
	for _, loc := range allLocs {
		selectors = append(selectors, &TimeSelector{
			Id:   loc.Id,
			Year: year,
			Week: week,
			Orus: loc.Orus[len(loc.Orus)-1],
		})
	}

	resa := GetSchedule(selectors...)

	var availableIds []string

	for _, res := range resa {
		lastUnderscore := strings.LastIndex(res.Id, "_")
		locationId := res.Id[lastUnderscore+1:]

		available := true
		for _, l1 := range res.Apps {
			StartTime, EndTime := l1.GetDates()
			if StartTime.Day() != sTime.Day() {
				continue
			}

			if (StartTime.Unix() > sTime.Unix() && StartTime.Unix() < eTime.Unix()) || // If search starts before and ends in a lesson
				(StartTime.Unix() < sTime.Unix() && EndTime.Unix() > eTime.Unix()) || // If search is entirely in a lesson
				(EndTime.Unix() > sTime.Unix() && EndTime.Unix() < eTime.Unix()) || // If search start in and ends after a lesson
				(StartTime.Unix() > sTime.Unix() && EndTime.Unix() < eTime.Unix()) { // If search is entirely over a lesson

				available = false
				break
			}
		}

		if available {
			availableIds = append(availableIds, locationId)
		}
	}

	var availableLocations []*Location
	for _, id := range availableIds {
		for _, loc := range allLocs {
			if loc.Id == id {
				availableLocations = append(availableLocations, loc)
				break
			}
		}

	}

	return availableLocations
}

func GetLocationById(id string) *Location {
	for _, location := range GetAllLocations() {
		if location.Id == id {
			return location
		}
	}
	return nil
}

//func GetAvailableLocations(prefix string, sTime time.Time, eTime time.Time) []*Location {
//	allLocs := GetAllLocationsWithPrefix(prefix, true)
//	var availableLocations []*Location
//
//	fmt.Println(len(allLocs))
//	t1 := time.Now().UnixMilli()
//
//	client := GetAndCheckCookies()
//	year, week := sTime.ISOWeek()
//	var allReady []bool
//
//	for _, l := range allLocs {
//		i := len(allReady)
//		allReady = append(allReady, false)
//		l := l
//		go func() {
//			//fmt.Println("Getting location " + l.Code)
//			get, err := client.Get("https://sa-curio.xedule.nl/api/schedule/?ids[0]=15_" + strconv.Itoa(year) + "_" + strconv.Itoa(week) + "_" + l.Id)
//			if err != nil {
//				fmt.Println(err)
//			} else {
//				da, _ := ioutil.ReadAll(get.Body)
//				//fmt.Println(string(da))
//				//d := json.NewDecoder(get.Body)
//
//				var resa []*XScheduleResponse
//				//err := d.Decode(&resa)
//				err := json.Unmarshal(da, &resa)
//				res := resa[0]
//				if err == nil {
//
//					available := true
//
//					for _, l1 := range res.Apps {
//						StartTime, EndTime := l1.GetDates()
//						if StartTime.Day() != sTime.Day() {
//							continue
//						}
//						if (StartTime.Unix() <= sTime.Unix() && EndTime.Unix() >= sTime.Unix()) || (StartTime.Unix() <= eTime.Unix() && EndTime.Unix() >= eTime.Unix()) {
//							available = false
//							break
//						}
//					}
//					if available {
//						availableLocations = append(availableLocations, l)
//					}
//				} else {
//					fmt.Println(l.Code)
//					fmt.Println(l.Code, err, string(da))
//					//d, _ := ioutil.ReadAll(*a)
//					//fmt.Println(string(d))
//					//fmt.Println(*res)
//					//fmt.Println(err)
//				}
//
//			}
//			allReady[i] = true
//
//		}()
//
//	}
//
//	notReady := true
//	for notReady {
//		notReady = false
//		for _, a := range allReady {
//			if !a {
//				notReady = true
//				break
//			}
//		}
//	}
//	t2 := time.Now().UnixMilli()
//	fmt.Println(t2 - t1)
//	return availableLocations
//}

func pullAllLocations() []*Location {
	client := GetAndCheckCookies()

	get, err := client.Get("https://sa-curio.xedule.nl/api/classroom/")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(get.StatusCode)

	if get.StatusCode != http.StatusOK {
		go Login()
		return Cache
	}

	var locations []*Location
	d := json.NewDecoder(get.Body)
	err = d.Decode(&locations)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	Cache = locations
	lastPulled1 := time.Now()
	lastPulled = &lastPulled1

	return locations
}
