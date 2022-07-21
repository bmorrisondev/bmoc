package services

import (
	"bmoc/cmd/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	bmoutils "github.com/bmorrisondev/go-utils"
	"github.com/gosimple/slug"
)

type WordPressClient struct {
	BaseUrl  string
	Username string
	Password string
}

func (c *WordPressClient) GetToken() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", c.Username, c.Password)))
}

// TODO: Better error handling
func (c *WordPressClient) UploadMediaFromUrl(url string, alt string, caption string) (*string, error) {
	// Setup
	fileDir, _ := os.Getwd()
	fileDir += "/tmp"
	urlsplit := strings.Split(url, "/")
	// Removes query params from the URL
	namesplit := strings.Split(urlsplit[len(urlsplit)-1], "?")
	filePath := path.Join(fileDir, namesplit[0])

	// Download the file into ./tmp
	err := utils.DownloadFile(url, filePath)
	if err != nil {
		return nil, err
	}
	newname := path.Join(fileDir, fmt.Sprintf("%v%v", slug.Make(alt), filepath.Ext(filePath)))
	err = os.Rename(filePath, newname)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(newname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create the POST body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	// Execute the request
	uploadUrl := fmt.Sprintf("%v/media", c.BaseUrl)
	hc := http.Client{}
	req, _ := http.NewRequest("POST", uploadUrl, body)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", c.GetToken()))
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Content-Disposition", fmt.Sprintf("attachment; filename=%v%v", slug.Make(alt), filepath.Ext(filePath)))
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var wpumr WordPressUploadMediaResponse
	err = json.Unmarshal(resb, &wpumr)
	if err != nil {
		return nil, err
	}

	// Update media with the necessary meta
	updateMediaReq := WordPressUpdateMediaRequest{
		AltText: alt,
		Caption: caption,
	}
	jstr, err := bmoutils.ConvertToJsonString(updateMediaReq)
	if err != nil {
		return nil, err
	}

	updateUrl := fmt.Sprintf("%v/%v", uploadUrl, wpumr.Id)
	req, err = http.NewRequest("POST", updateUrl, bytes.NewReader([]byte(jstr)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", c.GetToken()))
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/json")
	_, err = hc.Do(req)
	if err != nil {
		return nil, err
	}

	err = os.Remove(newname)
	if err != nil {
		return nil, err
	}

	return &wpumr.SourceUrl, nil
}

type WordPressUploadMediaResponse struct {
	Id        int    `json:"id"`
	SourceUrl string `json:"source_url"`
}

type WordPressUpdateMediaRequest struct {
	AltText string `json:"alt_text"`
	Caption string `json:"caption"`
}
