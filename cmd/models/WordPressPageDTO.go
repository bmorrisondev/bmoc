package models

type WordPressPageDTO struct {
	HTML           string
	Excerpt        string
	Title          string
	ImagesToUpload []WordPressMediaDTO
}

type WordPressMediaDTO struct {
	Name        string
	OriginalUrl string
	Slug        string
	Tag         string
}
