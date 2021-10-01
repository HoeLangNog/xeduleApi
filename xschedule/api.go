package xschedule

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type CachedCookie struct {
	Expires time.Time
	Cookie *http.Cookie
}

var CachedCookies []*CachedCookie

func init() {
	Login()
}

func Login() {
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
					"RelayState": relayState,
					"SAMLResponse": samlRes2,
				})

				CachedCookies = nil

				for _, c := range c4.Cookies("https://sa-curio.xedule.nl/") {
					newTime := time.Now()

					newTime.Add(time.Duration(time.Now().Unix()))
					cookie := &CachedCookie{
						Expires: newTime,
						Cookie: c,
					}

					CachedCookies = append(CachedCookies, cookie)

				}

				fmt.Println("Logged in!")


				//fmt.Println(string(e.Response.Body))
			})

			c3.Post(newPostUrl, map[string]string{
				"SAMLResponse": samlRes,
			})
		})

		c2.Post(postUrl, map[string]string{
			"UserName": os.Getenv("username"),
			"Password": os.Getenv("password"),
			"AuthMethod": "FormsAuthentication",
		})
	})

	c1.Visit("https://roosters.curio.nl")
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
