package client

import (
	"time"
)

type User struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	URL         string `json:"url"`
}

type Courses struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []*Course `json:"results"`
}

type Course struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	PublishedTitle string `json:"published_title"`
}

type PriceDetail struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	PriceString    string  `json:"price_string"`
	CurrencySymbol string  `json:"currency_symbol"`
}

type Lectures struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []*Lecture `json:"results"`
}

type Lecture struct {
	Chapter             *Chapter `json:"-"`
	ID                  int      `json:"id"`
	Title               string   `json:"title"`
	TitleCleaned        string   `json:"title_cleaned"`
	Asset               *Asset   `json:"asset"`
	SupplementaryAssets []*Asset `json:"supplementary_assets"`
	ObjectIndex         int      `json:"object_index"`
}

type Chapter struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ObjectIndex int    `json:"object_index"`
}

type Asset struct {
	ID           int           `json:"id"`
	AssetType    string        `json:"asset_type"`
	Title        string        `json:"title"`
	ExternalURL  string        `json:"external_url"`
	DownloadUrls *DownloadURLs `json:"download_urls"`
	//SlideUrls    []interface{} `json:"slide_urls"`
	StreamUrls *StreamURLs `json:"stream_urls"`
	Captions   []*Caption  `json:"captions"`
}

type DownloadURLs struct {
	Video []*Video `json:"Video"`
	File  []*File  `json:"File"`
}

type StreamURLs struct {
	Video []*Video `json:"Video"`
}

type Video struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	File  string `json:"file"`
}

type File struct {
	Label string `json:"label"`
	File  string `json:"file"`
}

type Caption struct {
	Status     int       `json:"status"`
	Locale     Locale    `json:"locale"`
	ID         int       `json:"id"`
	Source     string    `json:"source"`
	Title      string    `json:"title"`
	VideoLabel string    `json:"video_label"`
	Created    time.Time `json:"created"`
	FileName   string    `json:"file_name"`
	URL        string    `json:"url"`
}

type Locale struct {
	Locale string `json:"locale"`
}
