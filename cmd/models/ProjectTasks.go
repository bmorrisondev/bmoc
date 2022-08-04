package models

import "strings"

var creatorArticleTasks map[string][]string = map[string][]string{
	"Write draft":                     nil,
	"Featured image linked & created": nil,
	"Icon linked":                     nil,
	"GitHub set up": {
		"Readme created",
		"Website link added",
		"Add link to YouTube video",
	},
	"Draft uploaded to WordPress": nil,
	"Processed through Grammarly": nil,
	"Published article reviewed": {
		"Code is properly highlighted",
		"Files are bolded",
		"If the article is in a series, that info is properly populated as well",
		"Looks good on mobile & desktop viewports",
	},
	"Tweets written & scheduled": nil,
}

var creatorVideoTasks map[string][]string = map[string][]string{
	"Video recorded": {
		"Intro A",
		"Intro B",
		"Body",
		"Outro",
		"Turn off camera",
	},
	"Video edited": {
		"Trimmed",
		"Lower third",
		"Title card",
		"Background music",
		"YouTube sub/like/notify button",
		"Outro card",
		"Rendered",
	},
	"Video staged & uploaded": {
		"Title",
		"Tags",
		"Description",
		"Thumbnail",
		"First comment written & pinned",
		"End card setup properly",
	},
}

var creatorStreamTasks map[string][]string = map[string][]string{
	"Livestream planned & scheduled":  nil,
	"Livestream scheduled in YouTube": nil,
}

var planetscaleDocTasks map[string][]string = map[string][]string{
	"Write draft": nil,
	"Perform self review": {
		"Adheres to content guide",
		"Run through Grammarly",
		"All images have captions",
		"Interactable elements are bolded & quoted",
	},
	"Request review": {
		"Export to zip & process w/bmoc",
		"Created PR",
		"Tag Holly & legal",
	},
	"Cleanup old resources": nil,
	"Notify Jenn":           nil,
}

var planetscaleArticlesTasks map[string][]string = map[string][]string{
	"Write draft":              nil,
	"Featured image requested": nil,
	"Perform self review": {
		"Adheres to content guide",
		"Run through Grammarly",
		"All images have captions",
		"Interactable elements are bolded & quoted",
	},
	"Draft uploaded to GitHub, review PR created": nil,
}

func BuildProjectTaskList(area string, contentTypes []string) map[string][]string {
	projectTasks := map[string][]string{}
	if area == "Creator" {
		if contains(contentTypes, "Article") {
			for k, v := range creatorArticleTasks {
				projectTasks[k] = v
			}
		}

		if contains(contentTypes, "Video") {
			for k, v := range creatorVideoTasks {
				projectTasks[k] = v
			}
		}

		if contains(contentTypes, "Live Stream") {
			for k, v := range creatorStreamTasks {
				projectTasks[k] = v
			}
		}
	}

	if area == "PlanetScale" {
		if contains(contentTypes, "Doc") {
			for k, v := range creatorArticleTasks {
				projectTasks[k] = v
			}
		}

		if contains(contentTypes, "Article") {
			for k, v := range creatorArticleTasks {
				projectTasks[k] = v
			}
		}

	}
	return projectTasks
}

func contains(array []string, e string) bool {
	for _, el := range array {
		if strings.Contains(el, e) {
			return true
		}
	}
	return false
}
