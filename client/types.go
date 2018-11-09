package client

import "time"

type User struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	JobTitle     string `json:"job_title"`
	Image50X50   string `json:"image_50x50"`
	Image100X100 string `json:"image_100x100"`
	Initials     string `json:"initials"`
	URL          string `json:"url"`
}

type Courses struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []*Course `json:"results"`
}

type Course struct {
	ID                   int          `json:"id"`
	Title                string       `json:"title"`
	Description          string       `json:"description"`
	URL                  string       `json:"url"`
	IsPaid               bool         `json:"is_paid"`
	Price                string       `json:"price"`
	PriceDetail          PriceDetail  `json:"price_detail"`
	VisibleInstructors   []Instructor `json:"visible_instructors"`
	Image125H            string       `json:"image_125_H"`
	Image240X135         string       `json:"image_240x135"`
	IsPracticeTestCourse bool         `json:"is_practice_test_course"`
	Image480X270         string       `json:"image_480x270"`
	PublishedTitle       string       `json:"published_title"`
}

type PriceDetail struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	PriceString    string  `json:"price_string"`
	CurrencySymbol string  `json:"currency_symbol"`
}

type Instructor struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	JobTitle     string `json:"job_title"`
	Image50X50   string `json:"image_50x50"`
	Image100X100 string `json:"image_100x100"`
	Initials     string `json:"initials"`
	URL          string `json:"url"`
}

type Lectures struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []*Lecture `json:"results"`
}

type Lecture struct {
	ID                  int       `json:"id"`
	Title               string    `json:"title"`
	Created             time.Time `json:"created"`
	Description         string    `json:"description"`
	TitleCleaned        string    `json:"title_cleaned"`
	IsPublished         bool      `json:"is_published"`
	Transcript          string    `json:"transcript"`
	IsDownloadable      bool      `json:"is_downloadable"`
	IsFree              bool      `json:"is_free"`
	Asset               Asset     `json:"asset"`
	SupplementatyAssets []*Asset  `json:"supplementary_assets"`
	SortOrder           int       `json:"sort_order"`
	ObjectIndex         int       `json:"object_index"`
	Course              struct {
		ID     int    `json:"id"`
		URL    string `json:"url"`
		IsPaid bool   `json:"is_paid"`
	} `json:"course"`
}

type Asset struct {
	ID           int           `json:"id"`
	AssetType    string        `json:"asset_type"`
	Title        string        `json:"title"`
	Created      time.Time     `json:"created"`
	ExternalURL  string        `json:"external_url"`
	DownloadUrls *DownloadURLs `json:"download_urls"`
	SlideUrls    []interface{} `json:"slide_urls"`
	StreamUrls   *StreamURLs   `json:"stream_urls"`
	Captions     []*Caption    `json:"captions"`
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
