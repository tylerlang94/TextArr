package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	sonarrURL     = os.Getenv("SONARR_URL")
	sonarrAPI     = os.Getenv("SONARR_API")
	radarrURL     = os.Getenv("RADARR_URL")
	radarrAPI     = os.Getenv("RADARR_API")
	tvRootPath    = os.Getenv("TV_ROOT_PATH")
	movieRootPath = os.Getenv("MOVIE_ROOT_PATH")
)

func main() {
	// check config variables before proceeding
	// don't want to start server if any are null
	if sonarrAPI == "" || sonarrURL == "" {
		log.Fatal("Sonarr values are empty. Please fill them in before proceeding")
	}

	if radarrAPI == "" || radarrURL == "" {
		log.Fatal("Radarr values are empty. Please fill them in before proceeding")
	}

	// TODO: Set a default root path for TV and Movies if not set
	http.HandleFunc("/sms", smsHandler)

	port := "6000"
	log.Printf("listening on port %s... ", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func smsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing form:", err)
		http.Error(w, "Bad Requests:", http.StatusBadRequest)
		return
	}

	message := strings.TrimSpace(r.Form.Get("Body"))
	log.Println("Recieved message:", message)

	var responseText string

	if strings.HasPrefix(strings.ToLower(message), "request") {
		title := strings.TrimSpace(message[7:])
		if strings.Contains(strings.ToLower(message), "movies:") {
			if addToRadarr(title) {
				responseText = fmt.Sprintf("Movie '%s' added!", title)
			} else {
				responseText = fmt.Sprintf("Could not find movie '%s'", title)
			}
		} else {
			if addToSonarr(title) {
				responseText = fmt.Sprintf("Show '%s' added", title)
			} else {
				responseText = fmt.Sprintf("Could not find show '%s'", title)
			}
		}
	} else {
		responseText = "Use: Request <title> or Request movie: <title>"
	}

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprintf(w, "<Response><Message>%s</Message></Response>", responseText)
}

func addToRadarr(title string) bool {
	lookupUrl := fmt.Sprintf("%s/api/v3/movie/lookup?term=%s", radarrURL, url.QueryEscape(title))
	data, err := doGet(lookupUrl, radarrAPI)
	if err != nil {
		return false
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		return false
	}

	movie := results[0]
	payload := map[string]interface{}{
		"title":            movie["title"],
		"titleSlig":        movie["titleSlug"],
		"images":           movie["images"],
		"imdbId":           movie["tmdbId"],
		"year":             movie["year"],
		"rootFolderPath":   movieRootPath,
		"monitored":        true,
		"qualityProfileId": 1,
		"addOptions": map[string]bool{
			"searchForMovie": true,
		},
	}

	postUrl := fmt.Sprintf("%s/api/v3/movie", radarrURL)
	return doPost(postUrl, radarrAPI, payload)
}

func addToSonarr(title string) bool {
	lookupURL := fmt.Sprintf("%s/api/v3/series/lookup?term=%s", sonarrURL, url.QueryEscape(title))
	data, err := doGet(lookupURL, sonarrAPI)
	if err != nil {
		return false
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		return false
	}

	series := results[0]
	payload := map[string]interface{}{
		"title":            series["title"],
		"titleSlug":        series["titleSlug"],
		"images":           series["images"],
		"seasons":          series["seasons"],
		"rootFolderPath":   tvRootPath,
		"monitored":        true,
		"qualityProfileId": 1,
		"addOptions": map[string]bool{
			"searchForMissingEpisodes": true,
		},
	}

	postURL := fmt.Sprintf("%s/api/v3/series", sonarrURL)
	return doPost(postURL, sonarrAPI, payload)
}

func doGet(url, apiKey string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println("GET error: ", err)
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(req.Body)
}

func doPost(url, apiKey string, payload interface{}) bool {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("POST errors", err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	log.Println("POST failed with status code:", resp.StatusCode)
	return false
}
