package scraper

import (
	"github.com/channel-42/moodle-scraper/internal/courses"
	"github.com/channel-42/moodle-scraper/internal/login"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

func SetupCallbacks(c *colly.Collector) {
	c.OnRequest(func(r *colly.Request) {
		log.WithFields(log.Fields{"url": r.URL}).Debug("Visiting URL")
	})

	c.OnResponse(func(r *colly.Response) {
		log.WithFields(log.Fields{"code": r.StatusCode}).Debug("Got Response")
	})

	c.OnError(func(r *colly.Response, err error) {
		log.WithError(err).Debug("Request failed")
		log.Exit(1)
	})

	// login
	c.OnHTML(login.Body, login.Login)

	// verify login
	c.OnHTML(login.Body, login.VerifyLogin)

	// get all courses of user
	c.OnHTML(courses.Userprofile_element, courses.GetCourses)

	// get all resources of course
	c.OnHTML(courses.Coursesite_element, courses.GetCoursePdfs)

	// download resource of course
	c.OnResponse(courses.DownloadResource)
}

func LoginAndVerify(c *colly.Collector) {
	// visit login page twice to login and verify
	c.Visit(login.Url())
	c.Visit(login.Url())
}

func GetCourses(c *colly.Collector) {
	c.Visit(courses.Url())
}

func GetResources(c *colly.Collector, course *courses.Course) {
	c.Visit(course.Url)
}

func DownloadResource(c *colly.Collector, rr *[]courses.Resource) {
	for _, r := range *rr {
		log.WithFields(log.Fields{"name": r.Name}).Debug("Got resource selection")
		c.Visit(r.Url)
	}
}
