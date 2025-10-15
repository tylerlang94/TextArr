package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tylerlang94/TextArr/internal/configuration"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

/*
var (
	sonarrURL     = os.Getenv("SONARR_URL")
	sonarrAPI     = os.Getenv("SONARR_API")
	radarrURL     = os.Getenv("RADARR_URL")
	radarrAPI     = os.Getenv("RADARR_API")
	tvRootPath    = os.Getenv("TV_ROOT_PATH")
	movieRootPath = os.Getenv("MOVIE_ROOT_PATH")
)
*/

var configPath = flag.String("config", "", "Path to config file (YAML)")

type App struct {
	sonarrUrl     string
	sonarrApi     string
	radarrUrl     string
	radarrApi     string
	tvRootPath    string
	movieRootPath string
}

func main() {
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	app := &App{
		sonarrUrl:     cfg.Sonarr.URL,
		sonarrApi:     cfg.Radarr.API,
		radarrUrl:     cfg.Radarr.URL,
		radarrApi:     cfg.Radarr.API,
		tvRootPath:    cfg.Paths.TV,
		movieRootPath: cfg.Paths.Movies,
	}

	// check config variables before proceeding
	// don't want to start server if any are null
	if app.radarrApi == "" || app.radarrUrl == "" {
		log.Fatal("Sonarr values are empty. Please fill them in before proceeding")
	}

	if app.sonarrApi == "" || app.sonarrUrl == "" {
		log.Fatal("Radarr values are empty. Please fill them in before proceeding")
	}

	// TODO: Set a default root path for TV and Movies if not set
	http.HandleFunc("/sms", app.smsHandler)

	port := "6000"
	log.Printf("listening on port %s... ", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadConfig(path string) (*configuration.Config, error) {
	var cfg configuration.Config

	// Load YAML if provided
	if path != "" {
		// use your strict loader if you added it
		if err := configuration.LoadConfig(path, &cfg); err != nil {
			return nil, err
		}
	}

	// overlay environment and normalize/validate
	cfg.ApplyEnv()
	if err := cfg.Normalize(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (a *App) smsHandler(w http.ResponseWriter, r *http.Request) {
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
			if a.addToRadarr(title) {
				responseText = fmt.Sprintf("Movie '%s' added!", title)
			} else {
				responseText = fmt.Sprintf("Could not find movie '%s'", title)
			}
		} else {
			if a.addToSonarr(title) {
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

func (a *App) addToRadarr(title string) bool {
	lookupUrl := fmt.Sprintf("%s/api/v3/movie/lookup?term=%s", a.radarrUrl, url.QueryEscape(title))
	data, err := doGet(lookupUrl, a.radarrApi)
	if err != nil {
		return false
	}

	var results []map[string]any
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		return false
	}

	movie := results[0]
	payload := map[string]any{
		"title":            movie["title"],
		"titleSlig":        movie["titleSlug"],
		"images":           movie["images"],
		"imdbId":           movie["tmdbId"],
		"year":             movie["year"],
		"rootFolderPath":   a.movieRootPath,
		"monitored":        true,
		"qualityProfileId": 1,
		"addOptions": map[string]bool{
			"searchForMovie": true,
		},
	}

	postUrl := fmt.Sprintf("%s/api/v3/movie", a.radarrUrl)
	return doPost(postUrl, a.radarrApi, payload)
}

func (a *App) addToSonarr(title string) bool {
	lookupURL := fmt.Sprintf("%s/api/v3/series/lookup?term=%s", a.sonarrUrl, url.QueryEscape(title))
	data, err := doGet(lookupURL, a.sonarrApi)
	if err != nil {
		return false
	}

	var results []map[string]any
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		return false
	}

	series := results[0]
	payload := map[string]any{
		"title":            series["title"],
		"titleSlug":        series["titleSlug"],
		"images":           series["images"],
		"seasons":          series["seasons"],
		"rootFolderPath":   a.tvRootPath,
		"monitored":        true,
		"qualityProfileId": 1,
		"addOptions": map[string]bool{
			"searchForMissingEpisodes": true,
		},
	}

	postURL := fmt.Sprintf("%s/api/v3/series", a.sonarrUrl)
	return doPost(postURL, a.sonarrApi, payload)
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

func doPost(url, apiKey string, payload any) bool {
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
