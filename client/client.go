package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	HTTPClient  *http.Client
	Credentials Credentials
}

type Credentials struct {
	ID          string
	AccessToken string
}

type PaginationOptions struct {
	Page     int
	PageSize int
}

// LOGIN
const LoginFormURL = "https://www.udemy.com/join/login-popup/?display_type=popup&response_type=json"

const BaseURL = "https://www.udemy.com/api-2.0"
const (
	UserPath      = "users/me"
	MyCoursesPath = "users/me/subscribed-courses"
	CoursesPath   = "courses"
	Timeout       = time.Second * 600
)

func New() *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return &Client{
		HTTPClient: &http.Client{
			Jar:     jar,
			Timeout: Timeout,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}
}

func (c *Client) Login(ctx context.Context, email, password string) (Credentials, error) {
	var cred Credentials

	// load the form
	token, err := c.getCSRFToken(ctx)

	// Udemy is behind Cloudflare...
	time.Sleep(1 * time.Second)

	// prepare the request
	u := LoginFormURL
	params := url.Values{
		"email":               {email},
		"password":            {password},
		"csrfmiddlewaretoken": {token},
	}
	body := strings.NewReader(params.Encode())
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return cred, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Referer", "https://www.udemy.com/mobile/ipad/")
	req.Header.Set("Accept-Language", "jn-US;q=0.8,en;q=0.7")

	// call the backend
	_, err = c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return cred, err
	}

	loginURL, _ := url.Parse(LoginFormURL)
	for _, cookie := range c.HTTPClient.Jar.Cookies(loginURL) {
		if cookie.Name == "access_token" {
			cred.AccessToken = cookie.Value
		} else if cookie.Name == "client_id" {
			cred.ID = cookie.Value
		}
	}
	if cred.ID == "" || cred.AccessToken == "" {
		return cred, errors.New("could not load credentials from the response")
	}
	log.Printf("clientID=%s\n", cred.ID)
	log.Printf("accessToken=%s\n", cred.AccessToken)

	// Udemy is behind Cloudflare...
	time.Sleep(500 * time.Millisecond)

	c.Credentials = cred
	return cred, err
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
	defer func() {
		_ = res.Body.Close()
	}()
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
	if c.Credentials.ID != "" {
		req.Header.Set("X-Udemy-Client-Id", c.Credentials.ID)
	}
	if c.Credentials.AccessToken != "" {
		req.Header.Set("X-Udemy-Bearer-Token", c.Credentials.AccessToken)
		req.Header.Set("X-Udemy-Authorization", "Bearer "+c.Credentials.AccessToken)
		req.Header.Set("Authorization", "Bearer "+c.Credentials.AccessToken)
	}

	// and call the backend
	return c.HTTPClient.Do(req)
}

// Loads the login form and extracts the temporary CSRF token (used for login !)
func (c *Client) getCSRFToken(ctx context.Context) (string, error) {
	// load the HTML for the login form
	req, _ := http.NewRequest("GET", LoginFormURL, nil)
	// & add headers to avoid "robot detection"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Referer", "https://www.udemy.com/mobile/ipad/")
	req.Header.Set("Accept-Language", "jn-US;q=0.8,en;q=0.7")

	res, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err // could not contact the server
	}
	if res.StatusCode != 200 {
		_ = res.Body.Close()
		return "", fmt.Errorf("error loading the login form: status=%d", res.StatusCode) // server refused
	}

	// parse the HTML document
	// -> the CSRF token is held by an hidden element of the form
	//    <input type="hidden" name="csrfmiddlewaretoken" value="<TOKEN_IS_HERE>">
	doc, err := goquery.NewDocumentFromReader(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return "", err // un parsable HTML
	}
	input := doc.Find(".signin-form input[name=csrfmiddlewaretoken]") // <- find the hidden input
	if input.Length() == 0 {
		return "", errors.New("missing csrfmiddlewaretoken <input> element")
	}
	token, ok := input.Attr("value") // <- extract the value="..." from the input
	if !ok || token == "" {
		return "", errors.New("missing csrfmiddlewaretoken token value")
	}
	return token, nil
}
