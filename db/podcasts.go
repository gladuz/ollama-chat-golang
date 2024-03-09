package db

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type PodcastFeed struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Url            string `json:"url"`
	OriginalUrl    string `json:"originalUrl"`
	Link           string `json:"link"`
	Description    string `json:"description"`
	Author         string `json:"author"`
	Image          string `json:"image"`
	Artwork        string `json:"artwork"`
	LastUpdateTime int    `json:"lastUpdateTime"`
	Popularity     int    `json:"popularity"`
	EpisodeCount   int    `json:"episodeCount"`
	GUID           string `json:"guid"`
}

type PodSearchResult struct {
	Status string        `json:"status"`
	Feeds  []PodcastFeed `json:"feeds"`
	Count  int           `json:"count"`
}

type PodcastEpisode struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Link            string `json:"link"`
	Description     string `json:"description"`
	DatePublished   int    `json:"datePublished"`
	Duration        int    `json:"duration"`
	Episode         int    `json:"episode"`
	EnclosureUrl    string `json:"enclosureUrl"`
	EnclosureType   string `json:"enclosureType"`
	EnclosureLength int    `json:"enclosureLength"`
	GUID            string `json:"guid"`
}

type PodcastEpisodeResult struct {
	Status string         `json:"status"`
	Episode PodcastEpisode `json:"episode"`
}

type PodcastFeedResult struct {
	Items []PodcastEpisode `json:"items"`
	Count int              `json:"count"`
}

func PodcastOrgRequest(url string) []byte {
	err := godotenv.Load(".env.prod")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("PODCASTINDEX_API_KEY")
	apiSecret := os.Getenv("PODCASTINDEX_API_SECRET")

	now := time.Now()
	var apiHeaderTime string = strconv.FormatInt(now.Unix(), 10)
	var data4Hash string = apiKey + apiSecret + apiHeaderTime

	h := sha1.New()
	h.Write([]byte(data4Hash))
	hash := h.Sum(nil)
	hashString := fmt.Sprintf("%x", hash)

	// ======== Send the request and collect/show the results ========

	podClient := http.Client{
		Timeout: time.Second * 33,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "TestPodcastSearcher/0.1")
	req.Header.Set("X-Auth-Date", apiHeaderTime)
	req.Header.Set("X-Auth-Key", apiKey)
	req.Header.Set("Authorization", hashString)

	res, getErr := podClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr, "Error on request")

	}
	//parse the response

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	return body
}

func ParsePodcastSearchResult(query string) PodSearchResult {
	url := "https://api.podcastindex.org/api/1.0/search/byterm?q=" + query
	res := PodcastOrgRequest(url)
	// Parse json
	var result PodSearchResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result
}

func ParsePodcastEpisodesByShow(id string) PodcastFeedResult {
	url := "https://api.podcastindex.org/api/1.0/episodes/byfeedid?id=" + id + "&fulltext"
	res := PodcastOrgRequest(url)
	// Parse json
	var result PodcastFeedResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result
}

func ParseEpisode(id string) PodcastEpisode {
	url := "https://api.podcastindex.org/api/1.0/episodes/byid?fulltext&id=" + id
	res := PodcastOrgRequest(url)
	// Parse json
	var result PodcastEpisodeResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result.Episode
}