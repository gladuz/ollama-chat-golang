package models

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func podcastOrgRequest(url string) []byte {
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
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	return body
}

func ParsePodcastSearchResult(query string) PodsSearchResult {
	url := "https://api.podcastindex.org/api/1.0/search/byterm?q=" + query
	res := podcastOrgRequest(url)
	// Parse json
	var result PodsSearchResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result
}

func parsePodcastEpisodesByShow(id int) PodcastFeedResult {
	url := "https://api.podcastindex.org/api/1.0/episodes/byfeedid?id=" + strconv.Itoa(id) + "&fulltext"
	res := podcastOrgRequest(url)
	// Parse json
	var result PodcastFeedResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result
}

func parsePodcastFeed(id int) PodcastFeed {
	url := "https://api.podcastindex.org/api/1.0/podcasts/byfeedid?id=" + strconv.Itoa(id)
	res := podcastOrgRequest(url)
	// Parse json
	var result PodResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result.Feed
}

func parseEpisode(id int) PodcastEpisode {
	url := "https://api.podcastindex.org/api/1.0/episodes/byid?fulltext&id=" + strconv.Itoa(id)
	res := podcastOrgRequest(url)
	// Parse json
	var result PodcastEpisodeResult
	jsonErr := json.Unmarshal(res, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return result.Episode
}

func RefreshPodcastFeed(feedId int) {
	lastUpdateTime := 0
	err := DB.Get(&lastUpdateTime, "SELECT last_update_time FROM podcast_feed WHERE id = ?", feedId)
	if err == sql.ErrNoRows  {
		// if the feed is not in the database, add it
		pod := parsePodcastEpisodesByShow(feedId)
		for _, episode := range pod.Items {
			_, err = DB.NamedExec("INSERT INTO episode (id, title, link, description, date_published, duration, episode, enclosure_url, enclosure_type, enclosure_length, guid, feed_id) VALUES (:id, :title, :link, :description, :date_published, :duration, :episode, :enclosure_url, :enclosure_type, :enclosure_length, :guid, :feed_id)", episode)
			if err != nil {
				log.Fatal(err)
			}
		}
		return
	}
	//convert lastUpdateTime to time.Time
	lastUpdate := time.Unix(int64(lastUpdateTime), 0)
	if time.Since(lastUpdate).Hours() > 24 {
		fmt.Println("time to update")
		feed := parsePodcastEpisodesByShow(feedId)
		// put the feed in the database if the episode is not already there
		for _, episode := range feed.Items {
			//check if the episode is already in the database
			var temp_id int
			err := DB.Select(&temp_id, "SELECT id FROM episode WHERE id = ?", episode.ID)
			if err == sql.ErrNoRows {
				_, err_ex := DB.NamedExec("INSERT INTO episode (id, title, link, description, date_published, duration, episode, enclosure_url, enclosure_type, enclosure_length, guid, feed_id) VALUES (:id, :title, :link, :description, :date_published, :duration, :episode, :enclosure_url, :enclosure_type, :enclosure_length, :guid, :feed_id)", episode)
				if err_ex != nil {
					log.Fatal("err in inserting", err_ex)
				}
			}
		}

	}
}

func GetPodcastEpisodesByShow(feedId int) PodcastFeedResult {
	RefreshPodcastFeed(feedId)
	// Rest of the code...
	var episodes []PodcastEpisode
	DB.Select(&episodes, "SELECT * FROM episode WHERE feed_id = ?", feedId)		
	return PodcastFeedResult{Items: episodes, Count: len(episodes)}
}

func GetEpisode(id int) PodcastEpisode {
	var episode PodcastEpisode
	err := DB.Get(&episode, "SELECT * FROM episode WHERE id = ?", id)
	if err == nil{
		return episode
	}
	episode = parseEpisode(id)
	_, err = DB.NamedExec("INSERT INTO episode (id, title, link, description, date_published, duration, episode, enclosure_url, enclosure_type, enclosure_length, guid, feed_id) VALUES (:id, :title, :link, :description, :date_published, :duration, :episode, :enclosure_url, :enclosure_type, :enclosure_length, :guid, :feed_id)", episode)
	if err != nil {
		log.Fatal(err)
	}
	return episode
}

func GetPodcasts() []PodcastFeed {
	var podcasts []PodcastFeed
	err := DB.Select(&podcasts, "SELECT * FROM podcast_feed")
	if err != nil {
		log.Fatal(err)
	}
	return podcasts
}

func AddPodcast(podcastId int) {
	podcast := parsePodcastFeed(podcastId)
	_, err := DB.NamedExec("INSERT INTO podcast_feed (id, title, url, original_url, link, description, author, image, artwork, last_update_time, popularity, episode_count, guid) VALUES (:id, :title, :url, :original_url, :link, :description, :author, :image, :artwork, :last_update_time, :popularity, :episode_count, :guid)", podcast)
	if err != nil {
		log.Fatal(err)
	}
}