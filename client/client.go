package client

import (
	"context"
	"encoding/json"
	"fmt"
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
	Page     int
	PageSize int
}

const BaseURL = "https://www.udemy.com/api-2.0"
const (
	UserPath      = "users/me"
	MyCoursesPath = "users/me/subscribed-courses"
	CoursesPath   = "courses"
	Timeout       = time.Second * 600
)

func New() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: Timeout,
		},
	}
}

func (c *Client) Login(ctx context.Context, username, password string) (ID, accessToken string, err error) {
	return c.ID, c.AccessToken, nil
}

func (c *Client) GetUser(ctx context.Context) (*User, error) {
	var u *User
	err := c.getJson(ctx, BaseURL+"/"+UserPath, &u)
	return u, err
}

func (c *Client) ListCourses(ctx context.Context, opt *PaginationOptions) (*Courses, error) {
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
	err := c.getJson(ctx, u.String(), &cc)
	return cc, err
}

func (c *Client) GetCourse(ctx context.Context, ID int) (*Course, error) {
	u, _ := url.Parse(BaseURL)
	u.Path = path.Join(u.Path, MyCoursesPath, strconv.Itoa(ID))

	var course *Course
	err := c.getJson(ctx, u.String(), &course)
	return course, err
}

func (c *Client) LoadCurriculum(ctx context.Context, courseID int, opt *PaginationOptions) (*Curriculum, error) {
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

	// load curriculum as lectures
	var l *Curriculum
	err := c.getJson(ctx, u.String(), &l)
	return l, err
}

// getJSON calls GET and unmarshals the response JSON body
func (c *Client) getJson(ctx context.Context, url string, o interface{}) error {
	res, err := c.GET(ctx, url)
	if err != nil {
		return err
	}
	defer res.Body.Close() // won't fail !
	if res.StatusCode != 200 {
		// All calls to the API should response 200 OK
		return fmt.Errorf("failed call to Udemy: want .StatusCode=%d, got %d", 200, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(o)
}

// GET sends a GET request to Udemy, adding a whole lot of headers in the process
func (c *Client) GET(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	// taken from https://github.com/riazXrazor/udemy-dl/blob/master/lib/core.js
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:39.0) Gecko/20100101 Firefox/39.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Host", "www.udemy.com")
	req.Header.Set("Origin", "https://www.udemy.com")
	if c.ID != "" {
		req.Header.Set("X-Udemy-Client-Id", c.ID)
	}
	if c.AccessToken != "" {
		req.Header.Set("X-Udemy-Bearer-Token", c.AccessToken)
		req.Header.Set("X-Udemy-Authorization", "Bearer "+c.AccessToken)
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	// and call the backend
	return c.HTTPClient.Do(req)
}
