package controllers

import (
	"log"
	"net/http"
    "io/ioutil"

	"github.com/rss-creator/utils"
)

type ScraperController interface {
	GetWebsite(w http.ResponseWriter, r *http.Request)
}

type scraperController struct {
    Client *http.Client
}

func NewScraperController(client *http.Client) ScraperController {
	return &scraperController{client}
}

func (s *scraperController) GetWebsite(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		utils.SendError(w, "Url required", http.StatusBadRequest)
		return
	}

    resp, err := s.Client.Get(url)
    if err != nil {
        log.Printf("could not get response from url %v\n%v", url, err)
		utils.SendError(w, "Could not get response from url", http.StatusNotFound)
		return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("could not read response body from url %v\n%v", url, err)
		utils.SendError(w, "Could not read response from url", http.StatusBadRequest)
		return
    }

	utils.SendSuccess(w, string(body), http.StatusOK)
}
