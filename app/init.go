package app

import (
	"github.com/PROger4ever/go-simple-news-aggregator/app/models"
	rgorp "github.com/revel/modules/orm/gorp/app"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gorp.v2"
	"time"
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.ActionInvoker,           // Invoke the action.
	}
	revel.OnAppStart(func() {
		Dbm := rgorp.Db.Map
		setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
			for col, size := range colSizes {
				t.ColMap(col).MaxSize = size
			}
		}
		setColumnUniques := func(t *gorp.TableMap, colSizes map[string]bool) {
			for col, isUnique := range colSizes {
				t.ColMap(col).Unique = isUnique
			}
		}
		setColumnNotNull := func(t *gorp.TableMap, colSizes map[string]bool) {
			for col, isNotNull := range colSizes {
				t.ColMap(col).SetNotNull(isNotNull)
			}
		}

		t := Dbm.AddTable(models.User{}).SetKeys(true, "UserId")
		t.ColMap("Password").Transient = true
		setColumnSizes(t, map[string]int{
			"Username": 20,
			"Name":     100,
		})

		t = Dbm.AddTable(models.Source{}).SetKeys(true, "SourceId")
		setColumnUniques(t, map[string]bool{
			"Url": true,
		})
		setColumnSizes(t, map[string]int{
			"Name": 50,
			"Url":  200,
			"ArticlePageUrlsXpath": 1024,
			"CardXpath":            1024,
			"TitleXpath":           1024,
			"BodyXpath":            1024,
			"ImgXpath":             1024,
			"PublicationTimeXpath": 1024,
		})
		setColumnNotNull(t, map[string]bool{
			"Url": true,
			"ArticlePageUrlsXpath": false,
			"CardXpath":            false,
			"TitleXpath":           true,
			"BodyXpath":            true,
			"ImgXpath":             false,
			"PublicationTimeXpath": true,
		})

		t = Dbm.AddTable(models.Article{}).SetKeys(true, "ArticleId")
		t.ColMap("Source").Transient = true
		t.ColMap("PublicationTime").Transient = true
		t.SetUniqueTogether("SourceId", "PublicationTimeStr", "Title")
		setColumnSizes(t, map[string]int{
			"Url":    200,
			"Title":  200,
			"ImgUrl": 200,
		})
		setColumnNotNull(t, map[string]bool{
			"Title":              true,
			"Body":               true,
			"PublicationTimeStr": true,
		})

		rgorp.Db.TraceOn(revel.AppLog)
		Dbm.CreateTables()

		bcryptPassword, _ := bcrypt.GenerateFromPassword(
			[]byte("demo"), bcrypt.DefaultCost)
		demoUser := &models.User{
			Name:           "Demo User",
			Username:       "demo",
			Password:       "demo",
			HashedPassword: bcryptPassword,
		}
		if err := Dbm.Insert(demoUser); err != nil {
			panic(err)
		}

		sources := []*models.Source{
			//TODO: improve XPaths for Lenta.RU
			{
				SourceId:             0,
				Name:                 "Lenta.RU",
				Url:                  "https://lenta.ru/rubrics/russia/",
				ArticlePageUrlsXpath: `//*[@id="root"]/section[2]/div/div/div[1]/div/div[1]/section/div/div[2]/h3/a/@href`,
				CardXpath:            `//div[@itemtype="http://schema.org/NewsArticle"]`,
				TitleXpath:           `//div[@itemprop="headline"]`,
				BodyXpath:            `//div[@itemprop="articleBody"]`,
				ImgXpath:             `//img[@itemprop="url"]/@src`,
				PublicationTimeXpath: `//time[@itemprop="datePublished"]/@datetime`,
			},
		}
		for _, source := range sources {
			if err := Dbm.Insert(source); err != nil {
				panic(err)
			}
		}
		now := time.Now()
		articles := []*models.Article{
			{
				ArticleId: 0,
				SourceId:  0,

				Url:    "https://lenta.ru/news/2018/04/06/movement/",
				Body:   "It's a body example",
				ImgUrl: "//palacesquare.rambler.ru/ocwzjosl/MWF4eWE3LndhaHV1QHsiZGF0YSI6ey/JBY3Rpb24iOiJQcm94eSIsIlJlZmZl/cmVyIjoiaHR0cHM6Ly9sZW50YS5ydS/9uZXdzLzIwMTgvMDQvMDYvbW92ZW1l/bnQvIiwiUHJvdG9jb2wiOiJodHRwcz/oiLCJIb3N0IjoibGVudGEucnUifSwi/bGluayI6Imh0dHBzOi8vaWNkbi5sZW/50YS5ydS9pbWFnZXMvMjAxOC8wNC8w/Ni8xMS8yMDE4MDQwNjExNDQzNjgzNi/9waWNfODY4ZTQ2NTkwYmZkZTJlOTQx/MmViZjI5ZTg3YWVhMWQuanBnIn0%3D/",

				PublicationTimeStr: now.Format(models.SQL_DATE_FORMAT),

				PublicationTime: now,
				Source:          sources[0],
			},
		}
		for _, article := range articles {
			if err := Dbm.Insert(article); err != nil {
				panic(err)
			}
		}
	}, 5)
}

var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	// Add some common security headers
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}