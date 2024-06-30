package login

import (
	"github.com/channel-42/moodle-scraper/internal/config"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

var (
	Body          string = "body#page-login-index"
	error_element string = "a#loginerrormessage"
	logged_in     bool   = false
)

func Url() string {
	return config.Config.BaseUrl + "/login/index.php"
}

func Login(e *colly.HTMLElement) {
	if logged_in {
		return
	}

	action_url := e.Attr("action")

	// get unique login token
	login_token := e.ChildAttr("input[name=logintoken]", "value")

	login_data := make(map[string]string)
	login_data["username"] = config.User.Username
	login_data["password"] = config.User.Password
	login_data["logintoken"] = login_token

	log.Info("Logging in as ", login_data["username"])
	log.Debug("POST request to ", action_url)

	if err := e.Request.Post(action_url, login_data); err != nil {
		log.Error("Login failed", e)
	}
}

func VerifyLogin(e *colly.HTMLElement) {
	if logged_in {
		return
	}

	if e.ChildText(error_element) != "" {
		log.Fatal("Login failed")
		log.Exit(1)
	}

	logged_in = true
}
