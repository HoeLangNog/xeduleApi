package xschedule

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Response struct {
	Apps []*TimeSlot `json:"apps"`
	Id   string      `json:"id"`
}

type TimeSlot struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Summary    string `json:"summary"`
	Attention  string `json:"attention"`
	StartTimeS string `json:"iStart"`
	EndTimeS   string `json:"iEnd"`
	startTime  *time.Time
	endTime    *time.Time
	Attributes []int `json:"atts"`
}

func (timeSlot *TimeSlot) GetDates() (*time.Time, *time.Time) {
	if timeSlot.startTime == nil || timeSlot.endTime == nil {
		stime, _ := time.Parse("2006-01-02T15:04:05 MST", timeSlot.StartTimeS+" CET")
		etime, _ := time.Parse("2006-01-02T15:04:05 MST", timeSlot.EndTimeS+" CET")
		timeSlot.startTime = &stime
		timeSlot.endTime = &etime

	}
	return timeSlot.startTime, timeSlot.endTime
}

type TimeSelector struct {
	Orus int
	Year int
	Week int
	Id   string
}

var ScheduleCache = make(map[TimeSelector]*CachedSchedule)
var ScheduleCacheLock = &sync.RWMutex{}

type CachedSchedule struct {
	Schedule   *Response
	PulledTime time.Time
}

func GetSchedule(selectors ...*TimeSelector) []*Response {
	var schedulesInCache []*Response
	var toBeProcessed []*TimeSelector
	for _, selector := range selectors {
		selector.Year -= 1
		ScheduleCacheLock.RLock()
		cache, found := ScheduleCache[*selector]
		ScheduleCacheLock.RUnlock()
		if found {
			if cache.PulledTime.Unix() > time.Now().Unix()-1800 {
				schedulesInCache = append(schedulesInCache, cache.Schedule)
				continue
			}
		}

		toBeProcessed = append(toBeProcessed, selector)
	}
	chunks := chunkSchedules(toBeProcessed, 25)
	var responses []*Response
	responses = append(responses, schedulesInCache...)
	responsesLock := &sync.Mutex{}
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)

		chunk := chunk
		go func() {
			var localResponses []*Response
			query := "?"

			for i, selector := range chunk {
				query += "ids%5B" + strconv.Itoa(i) + "%5D=" + strconv.Itoa(selector.Orus) + "_" + strconv.Itoa(selector.Year) + "_" + strconv.Itoa(selector.Week) + "_" + selector.Id + "&"
			}

			query = query[:len(query)-1]

			if query != "" {
				client := GetAndCheckCookies()

				get, err := client.Get("https://sa-curio.xedule.nl/api/schedule" + query)

				if err != nil {
					fmt.Println(err)
					wg.Done()
					return
				}
				if get.StatusCode != http.StatusOK {
					fmt.Println(get.StatusCode)

					body, _ := ioutil.ReadAll(get.Body)
					fmt.Println(string(body))
					go func() {
						Login()
					}()

					wg.Done()
					return
				}

				d := json.NewDecoder(get.Body)

				err = d.Decode(&localResponses)

				if err != nil {
					fmt.Println(err)
					wg.Done()
					return
				}
			}

			for _, response := range localResponses {
				split := strings.Split(response.Id, "_")
				year, _ := strconv.Atoi(split[1])
				week, _ := strconv.Atoi(split[2])
				selector := TimeSelector{
					Id:   split[3],
					Year: year,
					Week: week,
				}
				ScheduleCacheLock.Lock()
				ScheduleCache[selector] = &CachedSchedule{
					Schedule:   response,
					PulledTime: time.Now(),
				}
				ScheduleCacheLock.Unlock()
			}

			responsesLock.Lock()
			responses = append(responses, localResponses...)
			responsesLock.Unlock()
			wg.Done()
		}()

	}

	wg.Wait()

	return responses
}

func chunkSchedules(selectors []*TimeSelector, size int) [][]*TimeSelector {
	var chunks [][]*TimeSelector
	for i := 0; i < len(selectors); i += size {
		end := i + size

		if end > len(selectors) {
			end = len(selectors)
		}

		chunks = append(chunks, selectors[i:end])
	}

	return chunks
}
