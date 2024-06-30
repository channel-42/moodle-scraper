package main

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/channel-42/moodle-scraper/internal/config"
	"github.com/channel-42/moodle-scraper/internal/courses"
	"github.com/channel-42/moodle-scraper/internal/scraper"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"
	"github.com/urfave/cli/v2"
)

var (
	s  *spinner.Spinner = spinner.New(spinner.CharSets[34], 100*time.Millisecond)
	ui *input.UI        = &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	cl *color.Color = color.New(color.FgHiGreen)
)

func main() {
	log.SetLevel(log.FatalLevel)

	app := &cli.App{
		Name:  "moodle-scraper",
		Usage: "Download resources from moodle",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"V"},
				Usage:   "Print version",
				Action: func(c *cli.Context, v bool) error {
					fmt.Println(config.Config.Version)
					log.Exit(0)
					return nil
				},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging",
				Action: func(c *cli.Context, v bool) error {
					if v {
						log.SetLevel(log.DebugLevel)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "loglevel",
				Usage:   "Set log level",
				Aliases: []string{"l"},
				Action: func(c *cli.Context, l string) error {
					level, err := log.ParseLevel(l)
					if err != nil {
						log.WithError(err).Fatal("Could not parse log level")
						return err
					}
					log.SetLevel(level)
					return nil
				},
			},
			&cli.BoolFlag{
				Name:    "all",
				Usage:   "Download all resources for a course",
				Aliases: []string{"a"},
				Action: func(c *cli.Context, a bool) error {
					config.Config.DownloadAll = a
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			return run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Exit(1)
	}
}

func run() error {
	// print banner
	cl.Println(config.Banner)

	// set credentials
	setCredentials()

	c := colly.NewCollector()

	scraper.SetupCallbacks(c)

	// try to login
	s.Start()
	scraper.LoginAndVerify(c)

	// get courses
	scraper.GetCourses(c)
	s.Stop()

	// debug print courses
	for _, course := range *courses.Courses {
		log.WithFields(log.Fields{
			"name": course.Name,
			"id":   course.Id,
			"url":  course.Url,
		}).Debug("Found course")
	}

	// get course selection
	courses_prompt := courses.BuildCourseSelectionPrompt()

	idx, _, err := courses_prompt.Run()

	if err != nil {
		log.WithError(err).Fatal("Prompt failed")
		return err
	}

	selected_course := &(*courses.Courses)[idx]

	// load course information
	s.Start()
	scraper.GetResources(c, selected_course)
	s.Stop()

	for _, r := range *selected_course.Resources {
		log.WithFields(log.Fields{
			"name": r.Name,
			"type": r.Type,
			"url":  r.Url,
		}).Debug("Found resource")
	}

	// check resource selction flag
	var selected_resources *[]courses.Resource
	if !config.Config.DownloadAll {
		// get resource selection
		selected_resources, err = courses.SelectResources(0, selected_course.Resources, selected_course)
		if err != nil {
			log.WithError(err).Fatal("Could not select resources")
			return err
		}
	} else {
		// select all all
		selected_resources = selected_course.Resources
	}

	// download resource
	s.Start()
	scraper.DownloadResource(c, selected_resources)
	s.Stop()

	cl.Println("Thank you for using moodle-scraper!")

	return nil
}

func setCredentials() {
	base_url, err := ui.Ask("Enter Moodle Base URL", &input.Options{
		Default:  config.Config.BaseUrl,
		Required: true,
	})
	if err != nil {
		log.WithError(err).Fatal("Could not read base url")
		log.Exit(1)
	}

	username, err := ui.Ask("Enter Username", &input.Options{
		Required: true,
	})
	if err != nil {
		log.WithError(err).Fatal("Could not read username")
		log.Exit(1)
	}

	password, err := ui.Ask("Enter Password", &input.Options{
		Required:    true,
		Mask:        true,
		MaskDefault: true,
	})
	if err != nil {
		log.WithError(err).Fatal("Could not read password")
		log.Exit(1)
	}

	config.Config.BaseUrl = base_url
	config.User.Username = username
	config.User.Password = password
}
