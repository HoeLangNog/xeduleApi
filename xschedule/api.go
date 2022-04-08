package xschedule

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

type CachedCookie struct {
	Expires time.Time
	Cookie  *http.Cookie
}

type Organization struct {
	Id string `json:"id"`
}

var Organizations []*Organization

var CachedCookies []*CachedCookie

var lastLogin *int64
var awaitingLogin bool = false

var loginLock = &sync.Mutex{}

func init() {
	go func() {
		fmt.Println("Trying to find teachersList.json")

		info, err := os.Stat("./teachersList.json")

		if err != nil {
			fmt.Println("Error finding teachersList.json")
			return
		}

		if info.IsDir() {
			fmt.Println("teachersList.json appears to be a directory (tf are you doing)")
			return
		}

		file, err := os.Open("./teachersList.json")

		if err != nil {
			fmt.Println("Failed the stat but couldn't open (might be permissions)")
			return
		}

		d := json.NewDecoder(file)

		err = d.Decode(&teacherNames)

		if err != nil {
			fmt.Println("Failed reading teacherList as a json")
			return
		}

		fmt.Println("Loaded names")
	}()

	Login()
}

func Login() {
	if !loginLock.TryLock() {
		return
	}

	if lastLogin != nil && *lastLogin > (time.Now().Unix()-82) {
		awaitingLogin = true

		timeToWait := (*lastLogin) - time.Now().Unix() + 82
		fmt.Println("waiting for login", timeToWait, "seconds")
		time.Sleep(time.Second * time.Duration(timeToWait))
	}
	lastLogina := time.Now().Unix()
	lastLogin = &lastLogina

	fmt.Println("Started logging in.")
	c1 := colly.NewCollector()

	c1.OnHTML("#loginForm", func(e *colly.HTMLElement) {
		newUrl := e.Attr("action")
		host := e.Request.URL.Host

		postUrl := "https://" + host + newUrl

		c2 := colly.NewCollector()

		c2.OnHTML("form", func(e *colly.HTMLElement) {
			newPostUrl := e.Attr("action")

			samlRes := e.ChildAttr("input", "value")

			c3 := colly.NewCollector()

			c3.OnHTML("#saml2Form", func(e *colly.HTMLElement) {
				newNewPostUrl := e.Attr("action")

				relayState := e.ChildAttr("input[name=\"RelayState\"]", "value")
				samlRes2 := e.ChildAttr("input[name=\"SAMLResponse\"]", "value")

				c4 := colly.NewCollector()

				c4.Post(newNewPostUrl, map[string]string{
					"RelayState":   relayState,
					"SAMLResponse": samlRes2,
				})

				CachedCookies = nil

				for _, c := range c4.Cookies("https://sa-curio.xedule.nl/") {
					newTime := time.Now()

					newTime.Add(time.Duration(time.Now().Unix()))
					cookie := &CachedCookie{
						Expires: newTime,
						Cookie:  c,
					}

					CachedCookies = append(CachedCookies, cookie)

				}

				fmt.Println("Logged in!")
				loginLock.Unlock()

				go func() {
					client := GetAndCheckCookies()

					res, err := client.Get("https://sa-curio.xedule.nl/api/organisationalUnit/")
					if err != nil {
						return
					}

					d := json.NewDecoder(res.Body)

					Organizations = nil
					err = d.Decode(&Organizations)
					if err != nil {
						return
					}
				}()

				//fmt.Println(string(e.Response.Body))
			})

			c3.Post(newPostUrl, map[string]string{
				"SAMLResponse": samlRes,
			})
		})

		c2.Post(postUrl, map[string]string{
			"UserName":   os.Getenv("username"),
			"Password":   os.Getenv("password"),
			"AuthMethod": "FormsAuthentication",
		})
	})

	c1.Visit("https://sa-curio.xedule.nl")
}

func GetAndCheckCookies() *http.Client {
	client := &http.Client{}
	u, _ := url.Parse("https://sa-curio.xedule.nl/")
	var cookies []*http.Cookie
	for _, cachedCookie := range CachedCookies {
		//if time.Now().Unix() > cachedCookie.Expires.Unix()  {
		//	Login()
		//	return GetAndCheckCookies()
		//}
		cookies = append(cookies, cachedCookie.Cookie)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	jar.SetCookies(u, cookies)
	client.Jar = jar
	return client
}

func OrganizationIds() []int {
	var ids []int
	for _, organization := range Organizations {
		id, err := strconv.Atoi(organization.Id)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}
