package courses

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/channel-42/moodle-scraper/internal/config"
	"github.com/gocolly/colly/v2"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
)

type Resource struct {
	Name     string
	Type     string
	Size     float64
	Unit     string
	Url      string
	Selected bool
}

type Course struct {
	Name      string
	Id        int
	Url       string
	Resources *[]Resource
}

var (
	Courses                         *[]Course = &[]Course{}
	Userprofile_element             string    = "div.userprofile"
	Coursesite_element              string    = "div.course-content"
	profile_tree_block_query        string    = "section.node_category>div.card-body"
	courses_list_query              string    = "ul>li>dl>dd>ul>li"
	course_activity_item_query      string    = "div.activity-item"
	course_activity_item_link_query string    = "div.activityname>a"
)

func Url() string {
	return config.Config.BaseUrl + "/user/profile.php"
}

func CourseBaseUrl() string {
	return config.Config.BaseUrl + "/course/view.php?id="
}

func GetCourseByName(name string) *Course {
	idx := slices.IndexFunc(*Courses, func(c Course) bool {
		return c.Name == name
	})
	return &(*Courses)[idx]
}

func GetResourceByName(course *Course, name string) *Resource {
	idx := slices.IndexFunc(*course.Resources, func(r Resource) bool {
		return r.Name == name
	})
	return &(*course.Resources)[idx]
}

func BuildCourseSelectionPrompt() promptui.Select {
	return promptui.Select{
		Label: "Select course",
		Items: *Courses,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\u27A4 {{ .Name | cyan }} (ID {{ .Id | red }})",
			Inactive: "  {{ .Name | cyan }} (ID {{ .Id | red }})",
      Selected: " ",
		},
	}
}

func countSelectedResources(resources *[]Resource) int {
	count := 0
	for _, resource := range *resources {
		if resource.Selected {
			count++
		}
	}
	return count
}

func SelectResources(idx int, resources *[]Resource, course *Course) (*[]Resource, error) {
	// Always prepend a "Done" item to the slice if it doesn't
	// already exist.
	const done = "Done"
	if len(*resources) > 0 && (*resources)[0].Name != done {
		var items = &[]Resource{
			{
				Name: done,
				Unit: "Selected",
			},
		}
		*resources = append(*items, *resources...)
	}

	prompt := promptui.Select{
		Label: "Select resource to download",
		Items: *course.Resources,
		Templates: &promptui.SelectTemplates{
			Label: `{{if .Selected}}
                    ✔
                {{end}}{{ . }}?`,
			Active:   "\u27A4 {{if .Selected}}✔{{end}} {{ .Name | cyan }} ({{ .Type | red }} {{ .Size | blue }} {{ .Unit | blue }})",
			Inactive: "{{if .Selected}}✔{{end}} {{ .Name | cyan }} ({{ .Type | red }} {{ .Size | blue }} {{ .Unit | blue }})",
		},
		CursorPos:    idx,
		HideSelected: true,
	}

	selection_idx, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	if (*resources)[selection_idx].Name != done {
		//Directly toggle Selected property
		(*resources)[selection_idx].Selected = !(*resources)[selection_idx].Selected
    
    // update selction count
		d := GetResourceByName(course, done)
		d.Size = float64(countSelectedResources(resources))

		return SelectResources(selection_idx, resources, course)
	}

	// If the user selected the "Done" item, return
	// all selected items.
	selections := &[]Resource{}
	for _, i := range *resources {
		if i.Selected {
			*selections = append(*selections, i)
		}
	}
	return selections, nil
}

func GetCourses(e *colly.HTMLElement) {

	log.Debug("Found profile page")

	// iterate profile cards
	e.ForEach(profile_tree_block_query, func(i int, card *colly.HTMLElement) {

		// course list found
		if card.ChildText("h3") == "Kursdetails" {

			log.Debug("Found course overview")

			// iterate courses
			card.ForEach(courses_list_query, func(i int, course *colly.HTMLElement) {

				log.Debug("Found course")

				// parse course link
				parse_url, err := url.Parse(course.ChildAttr("a", "href"))
				if err != nil {
					log.Error("Error parsing URL:", err)
					return
				}

				// convert course id to int
				course_id, err := strconv.Atoi(parse_url.Query().Get("course"))
				if err != nil {
					log.Error("Error getting course id:", err)
					return
				}

				// create and append course
				c := Course{
					Name: course.ChildText("a"),
					Url:  CourseBaseUrl() + strconv.Itoa(course_id),
					Id:   course_id,
				}
				*Courses = append(*Courses, c)
			})
		}
	})

}

func GetCoursePdfs(e *colly.HTMLElement) {

	requested_course_id, err := strconv.Atoi(e.Request.URL.Query().Get("id"))
	if err != nil {
		log.Error("Error getting course id:", err)
		return
	}

	log.Debug("Searching for pdfs in course")

	rr := &[]Resource{}

	e.ForEach(course_activity_item_query, func(i int, item *colly.HTMLElement) {
		resource_details := item.ChildText("span.resourcelinkdetails")

		if resource_details == "" {
			log.Warn("Empty resource details")
			return
		}

		r, err := parseAndMakeResource(resource_details)

		if err != nil {
			log.Error("Error parsing resource ", err)
			return
		}

		r.Url = item.ChildAttr(course_activity_item_link_query, "href") + "&redirect=1"
		r.Name = item.ChildText("span.instancename")
		r.Selected = false
		*rr = append(*rr, *r)
	})

	// find requested course
	idx := slices.IndexFunc(*Courses, func(c Course) bool {
		return c.Id == requested_course_id
	})

	(*Courses)[idx].Resources = rr
}

func DownloadResource(r *colly.Response) {
	// check if the content type is a PDF
	content_type := r.Headers.Get("Content-Type")
	if !strings.Contains(content_type, "application/pdf") {
		return
	}

	// extract the filename from the URL
	filename := path.Base(r.Request.URL.Path)
	out, err := os.Create("./" + filename)
	if err != nil {
		log.WithError(err).Error("Could not create file")
		return
	}
	defer out.Close()

	log.WithFields(log.Fields{"url": r.Request.URL.String()}).Debug("Downloading resource")

	// write the body to file
	_, err = io.Copy(out, bytes.NewReader(r.Body))
	if err != nil {
		log.WithError(err).Error("Could not write file")
		return
	}
}

func parseAndMakeResource(text string) (*Resource, error) {
	// Normalize the input by replacing non-breaking spaces with regular spaces
	normalizedText := strings.ReplaceAll(text, "\u00a0", " ")

	// Apply the regular expression on the normalized text
	re := regexp.MustCompile(`(\d+(\.\d+)?)\s*(KB|MB)\s*(.+)`)
	matches := re.FindStringSubmatch(normalizedText)

	if matches == nil || len(matches) < 5 {
		return nil, fmt.Errorf("no matches found")
	}

	// Extract the size as float64
	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing size: %v", err)
	}

	// Extract the unit and type
	unit := matches[3]
	resourceType := matches[4]

	return &Resource{
		Size: size,
		Unit: unit,
		Type: resourceType,
	}, nil
}
