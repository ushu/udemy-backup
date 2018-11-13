package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const DefaultPageSize = 1400

type Client struct {
	ID          string
	AccessToken string
	HTTPClient  *http.Client
}

type PaginationOptions struct {
	Page     int
	PageSize int
}

const BaseURL = "https://www.udemy.com/api-2.0"
const (
	UserPath      = "users/me"
	MyCoursesPath = "users/me/subscribed-courses"
	CoursesPath   = "courses"
)

func New(id, accessToken string) *Client {
	// fix issue when udemy asset servers are too slow... (useful with high concurrency, typically ≧ 8)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &Client{
		ID:          id,
		AccessToken: accessToken,
		HTTPClient: &http.Client{
			Transport: tr,
			Timeout:   time.Second * 600,
		},
	}
}

func (c *Client) GetUser() (*User, error) {
	var u *User
	err := c.GetJson(BaseURL+"/"+UserPath, &u)
	return u, err
}

func (c *Client) ListCourses(opt *PaginationOptions) (*Courses, error) {
	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, MyCoursesPath)
	// add page info
	q := u.Query()
	q.Set("fields[course]", "@min,title,published_title")
	if opt != nil {
		if opt.Page > 1 {
			q.Set("page", strconv.Itoa(opt.Page))
		}
		if opt.PageSize > 1 {
			q.Set("page_size", strconv.Itoa(opt.PageSize))
		}
	}
	u.RawQuery = q.Encode()

	var cc *Courses
	err := c.GetJson(u.String(), &cc)
	return cc, err
}

func (c *Client) ListAllCourses() ([]*Course, error) {
	var cc []*Course
	opt := &PaginationOptions{
		Page:     1,
		PageSize: DefaultPageSize,
	}
	for {
		// load page info
		res, err := c.ListCourses(opt)
		if err != nil {
			return cc, err
		}
		cc = append(cc, res.Results...)

		// last page ?
		if res.Next == "" {
			break
		}
		opt.Page++
	}
	return cc, nil
}

func (c *Client) GetCourse(ID int) (*Course, error) {
	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, MyCoursesPath, strconv.Itoa(ID))

	var course *Course
	err := c.GetJson(u.String(), &course)
	return course, err
}

func (c *Client) LoadCurriculum(courseID int, opt *PaginationOptions) (*Curriculum, error) {
	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, CoursesPath, strconv.Itoa(courseID), "cached-subscriber-curriculum-items")
	q := u.Query()
	q.Set("fields[asset]", "@min,download_urls,stream_urls,external_url,slide_urls,captions")
	q.Set("fields[lecture]", "@min,title,title_cleaned,asset,object_index,supplementary_assets")
	q.Set("fields[caption]", "@min,file_name,locale,url")
	q.Set("fields[chapter]", "@min,title,object_index")
	if opt != nil {
		if opt.Page > 1 {
			q.Set("page", strconv.Itoa(opt.Page))
		}
		if opt.PageSize > 1 {
			q.Set("page_size", strconv.Itoa(opt.PageSize))
		}
	}
	u.RawQuery = q.Encode()

	// load curricullum as lectures
	var l Lectures
	err := c.GetJson(u.String(), &l)
	if err != nil {
		return nil, err
	}

	// some of the loaded "lectures", are in fact chapters !
	// (but we still load all as Lecture since chapters as similar keys)
	var currentChapter *Chapter
	var results []*Lecture
	for _, lc := range l.Results {
		switch lc.Class {
		case "chapter":
			currentChapter = &Chapter{
				ID:          lc.ID,
				Title:       lc.Title,
				ObjectIndex: lc.ObjectIndex,
			}
		case "lecture":
			lc.Chapter = currentChapter
			results = append(results, lc)
		}
	}

	return &Curriculum{
		Count:    len(results),
		Next:     l.Next,
		Previous: l.Previous,
		Results:  results,
	}, nil
}

func (c *Client) LoadFullCurriculum(courseID int) ([]*Lecture, error) {
	var res []*Lecture
	opt := &PaginationOptions{
		Page:     1,
		PageSize: DefaultPageSize,
	}
	for {
		// load page info
		cur, err := c.LoadCurriculum(courseID, opt)
		if err != nil {
			return res, err
		}
		res = append(res, cur.Results...)

		// last page ?
		if cur.Next == "" {
			break
		}
		opt.Page++
	}

	return res, nil
}

func (c *Client) GetJson(url string, o interface{}) error {
	// call the API
	res, err := c.GET(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// all Udemy API call should return 200 OK
		return fmt.Errorf("call to Udemy failed with status: %s", res.Status)
	}

	// read complete body
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// and decode the result
	return json.Unmarshal(b, o)
}

func (c *Client) GET(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// taken from https://github.com/riazXrazor/udemy-dl/blob/master/lib/core.js
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:39.0) Gecko/20100101 Firefox/39.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Host", "www.udemy.com")
	req.Header.Set("Origin", "https://www.udemy.com")
	req.Header.Set("X-Udemy-Bearer-Token", c.AccessToken)
	req.Header.Set("X-Udemy-Client-Id", c.ID)
	req.Header.Set("X-Udemy-Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// and call the backend
	return c.Do(req)
}

func (c *Client) RawGET(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	h := c.HTTPClient
	if h == nil {
		h = http.DefaultClient
	}
	return h.Do(req)
}
