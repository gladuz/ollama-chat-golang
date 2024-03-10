package models

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

type PodcastFeed struct {
	ID             int    `json:"id" db:"id"`
	Title          string `json:"title" db:"title"`
	Url            string `json:"url" db:"url"`
	OriginalUrl    string `json:"originalUrl" db:"original_url"`
	Link           string `json:"link" db:"link"`
	Description    string `json:"description" db:"description"`
	Author         string `json:"author" db:"author"`
	Image          string `json:"image" db:"image"`
	Artwork        string `json:"artwork" db:"artwork"`
	LastUpdateTime int    `json:"lastUpdateTime" db:"last_update_time"`
	Popularity     int    `json:"popularity" db:"popularity"`
	EpisodeCount   int    `json:"episodeCount" db:"episode_count"`
	GUID           string `json:"guid" db:"guid"`
}

type PodcastEpisode struct {
	ID              int    `json:"id" db:"id"`
	Title           string `json:"title" db:"title"`
	Link            string `json:"link" db:"link"`
	Description     string `json:"description" db:"description"`
	DatePublished   int    `json:"datePublished" db:"date_published"`
	Duration        int    `json:"duration" db:"duration"`
	Episode         int    `json:"episode" db:"episode"`
	EnclosureUrl    string `json:"enclosureUrl" db:"enclosure_url"`
	EnclosureType   string `json:"enclosureType" db:"enclosure_type"`
	EnclosureLength int    `json:"enclosureLength" db:"enclosure_length"`
	GUID            string `json:"guid" db:"guid"`
	FeedID          int    `json:"feedId" db:"feed_id"`
}

var scheme = `
CREATE TABLE IF NOT EXISTS podcast_feed (
	id INT,
	title VARCHAR(255) NOT NULL,
	url VARCHAR(255) NOT NULL,
	original_url VARCHAR(255) NOT NULL,
	link VARCHAR(255) NOT NULL,
	description TEXT NOT NULL,
	author VARCHAR(255) NOT NULL,
	image VARCHAR(255) NOT NULL,
	artwork VARCHAR(255) NOT NULL,
	last_update_time INT NOT NULL,
	popularity INT NOT NULL,
	episode_count INT NOT NULL,
	guid VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS episode (
	id INT,
	title VARCHAR(255) NOT NULL,
	link VARCHAR(255) NOT NULL,
	description TEXT NOT NULL,
	date_published INT NOT NULL,
	duration INT NOT NULL,
	episode INT NOT NULL,
	enclosure_url VARCHAR(255) NOT NULL,
	enclosure_type VARCHAR(255) NOT NULL,
	enclosure_length INT NOT NULL,
	guid VARCHAR(255) NOT NULL,
	feed_id INT NOT NULL,
	FOREIGN KEY (feed_id) REFERENCES podcast_feed(id)
);
`

type PodsSearchResult struct {
	Status string        `json:"status"`
	Feeds  []PodcastFeed `json:"feeds"`
	Count  int           `json:"count"`
}

type PodResult struct {
	Status string      `json:"status"`
	Feed   PodcastFeed `json:"feed"`
}

type PodcastEpisodeResult struct {
	Status  string         `json:"status"`
	Episode PodcastEpisode `json:"episode"`
}

type PodcastFeedResult struct {
	Items []PodcastEpisode `json:"items"`
	Count int              `json:"count"`
}

func InitDatabase() {
	var err error
	DB, err = sqlx.Connect("sqlite3", "./podllama.db")
	if err != nil {
		panic(err)
	}
	DB.MustExec(scheme)
}