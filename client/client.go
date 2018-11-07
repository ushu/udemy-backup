package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Client struct {
	ID          string
	AccessToken string
	HTTPClient  *http.Client
}

type PaginationOptions struct {
	Page int
}

const BaseURL = "https://www.udemy.com/api-2.0"
const (
	UserPath    = "users/me"
	CoursesPath = "users/me/subscribed-courses"
)

func New(id, accessToken string) *Client {
	return &Client{
		ID:          id,
		AccessToken: accessToken,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
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
	u.Path = path.Join(u.Path, CoursesPath)
	// add page info
	q := u.Query()
	q.Set("page_size", "1400")
	q.Set("fields[course]", "@default,description")
	if opt != nil && opt.Page > 1 {
		q.Set("page", strconv.Itoa(opt.Page))
	}
	u.RawQuery = q.Encode()

	var cc *Courses
	err := c.GetJson(u.String(), &cc)
	return cc, err
}

func (c *Client) ListAllCourses() ([]Course, error) {
	var cc []Course
	opt := &PaginationOptions{
		Page: 1,
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
	u.Path = path.Join(u.Path, CoursesPath, strconv.Itoa(ID))

	var course *Course
	err := c.GetJson(u.String(), &course)
	return course, err
}

func (c *Client) ListLectures(courseID int, opt *PaginationOptions, details bool) (Lectures, error) {
	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, CoursesPath, strconv.Itoa(courseID), "lectures")
	q := u.Query()
	q.Set("page_size", "1400")
	if details {
		q.Set("fields[asset]", "@min,download_urls,stream_urls,external_url,slide_urls,captions")
		q.Set("fields[lecture]", "@default,view_html,course,object_index,supplementary_assets")
	}
	if opt != nil && opt.Page > 1 {
		// add page info
		q.Set("page", strconv.Itoa(opt.Page))
	}
	u.RawQuery = q.Encode()

	var l Lectures
	err := c.GetJson(u.String(), &l)
	return l, err
}

func (c *Client) ListAllLectures(courseID int, details bool) ([]*Lecture, error) {
	var cc []*Lecture
	opt := &PaginationOptions{
		Page: 1,
	}
	for {
		// load page info
		res, err := c.ListLectures(courseID, opt, details)
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

	//finally we sort by sort order
	//sort.Slice(cc, func(i, j int) bool {
	//	return cc[i].SortOrder < cc[j].SortOrder
	//})

	return cc, nil
}

func (c *Client) GetLecture(courseID, lectureID int) (*Lecture, error) {
	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, CoursesPath, strconv.Itoa(courseID), "lectures", strconv.Itoa(lectureID))
	// add field options
	q := u.Query()
	q.Set("fields[asset]", "download_urls,stream_urls,external_url,slide_urls,captions")
	u.RawQuery = q.Encode()

	var l *Lecture
	err := c.GetJson(u.String(), &l)
	return l, err
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
