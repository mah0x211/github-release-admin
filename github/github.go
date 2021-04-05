package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github-release-admin/log"
)

type Client struct {
	baseURL    string
	uploadURL  string
	baseHeader http.Header
	Header     http.Header
	Body       io.Reader
}

func resolveURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return u.ResolveReference(&url.URL{}).String(), nil
}

var ReOwnerName = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
var ReRepoName = regexp.MustCompile(`^[\w-][\w.-]*$`)

func New(repo string) (*Client, error) {
	if repo = strings.TrimSpace(repo); repo == "" {
		return nil, fmt.Errorf("repo name must not be empty")
	}

	arr := strings.Split(repo, "/")
	if len(arr) != 2 {
		return nil, fmt.Errorf("invalid repo name %q", repo)
	}
	// verify ownername
	s := arr[0]
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") || !ReOwnerName.MatchString(s) {
		return nil, fmt.Errorf("invalid repo name %q", repo)
	}
	s = arr[1]
	if !ReRepoName.MatchString(s) {
		return nil, fmt.Errorf("invalid repo name %q", repo)
	}

	c := &Client{
		baseURL: "https://api.github.com/repos/" + repo,
		// uploadURL: "http://127.0.0.1:10080/",
		uploadURL: "https://uploads.github.com/repos/" + repo + "/releases",
		baseHeader: http.Header{
			"Accept": {"application/vnd.github.v3+json"},
		},
	}

	// use GITHUB_TOKEN if define
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		c.baseHeader.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	return c, nil
}

func (c *Client) log(req *http.Request, body bool) error {
	if log.Verbose {
		b, err := httputil.DumpRequest(req, body)
		if err != nil {
			return err
		}
		log.Printf("%s", b)
	}
	return nil
}

func (c *Client) createRequest(method, url string) (*http.Request, error) {
	body := c.Body
	header := c.Header
	c.Body = nil
	c.Header = nil

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = c.baseHeader.Clone()
	for k, v := range header {
		list, ok := req.Header[k]
		if !ok {
			list = make([]string, 0, len(v))
			copy(list, v)
		} else {
			list = append(list, v...)
		}
		req.Header[k] = list
	}

	return req, nil
}

func (c *Client) request(method, endpoint string) (*http.Response, error) {
	u, err := resolveURL(endpoint)
	if err != nil {
		return nil, err
	}

	req, err := c.createRequest(method, c.baseURL+u)
	if err != nil {
		return nil, err
	} else if err = c.log(req, true); err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) Get(endpoint string) (*http.Response, error) {
	return c.request("GET", endpoint)
}

func (c *Client) Post(endpoint string) (*http.Response, error) {
	return c.request("POST", endpoint)
}

func (c *Client) Delete(endpoint string) (*http.Response, error) {
	return c.request("DELETE", endpoint)
}

type Branch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

func (c *Client) GetBranch(name string) (*Branch, error) {
	rsp, err := c.Get(fmt.Sprintf("/branches/%s", name))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusOK:
		branch := &Branch{}
		if err := json.NewDecoder(rsp.Body).Decode(&branch); err != nil {
			return nil, err
		}
		return branch, nil

	case http.StatusNotFound:
		return nil, nil

	default:
		b, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return nil, err
	}
}

type Author struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HtmlURL   string `json:"html_url"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

type Asset struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

type Release struct {
	ID              int     `json:"id"`
	Draft           bool    `json:"draft"`
	PreRelease      bool    `json:"prerelease"`
	Name            string  `json:"name"`
	Body            string  `json:"body"`
	TagName         string  `json:"tag_name"`
	TargetCommitish string  `json:"target_commitish"`
	HtmlURL         string  `json:"html_url,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	PublishedAt     string  `json:"published_at,omitempty"`
	Author          Author  `json:"author,omitempty"`
	Assets          []Asset `json:"assets,omitempty"`
}

type ListRelease struct {
	NextPage int
	Releases []*Release
}

// Link: <https://api.github.com/repositories/194783954/releases?per_page=1&page=2>; rel=\"next\",
var ReLinkNext = regexp.MustCompile(`<([^>]+)>; rel="next"`)

func (c *Client) getNextPage(link string) (int, error) {
	if m := ReLinkNext.FindStringSubmatch(link); len(m) == 0 {
		return 0, nil
	} else if u, err := url.Parse(m[1]); err != nil {
		return 0, fmt.Errorf("invalid url: %w", err)
	} else if page := strings.TrimSpace(u.Query().Get("page")); page == "" {
		return 0, fmt.Errorf("page query not defined")
	} else if iv, err := strconv.Atoi(page); err != nil {
		return 0, fmt.Errorf("invalid page query: %w", err)
	} else {
		return iv, nil
	}
}

func (c *Client) ListReleases(nitem, page int) (*ListRelease, error) {
	rsp, err := c.Get(fmt.Sprintf("/releases?per_page=%d&page=%d", nitem, page))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNotFound:
		return nil, nil

	case http.StatusOK:
		list := &ListRelease{}
		if err = json.NewDecoder(rsp.Body).Decode(&list.Releases); err != nil {
			return nil, err
		}

		for _, v := range rsp.Header.Values("Link") {
			if page, err = c.getNextPage(v); err != nil {
				log.Errorf("faild to Link header: %v", err)
			} else {
				list.NextPage = page
			}
		}

		return list, nil

	default:
		b, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return nil, err
	}
}

func (c *Client) CreateRelease(tagName, targetCommitish, name, body string, draft, prerelease bool) (*Release, error) {
	b, err := json.Marshal(&Release{
		TagName:         tagName,
		TargetCommitish: targetCommitish,
		Name:            name,
		Body:            body,
		Draft:           draft,
		PreRelease:      prerelease,
	})
	if err != nil {
		return nil, err
	}

	c.Body = bytes.NewBuffer(b)
	rsp, err := c.Post("/releases")
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusCreated:
		release := &Release{}
		if err := json.NewDecoder(rsp.Body).Decode(&release); err != nil {
			return nil, err
		}
		return release, nil

	default:
		b, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return nil, err
	}
}

func (c *Client) DeleteRelease(id int) error {
	rsp, err := c.Delete(fmt.Sprintf("/releases/%d", id))
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusNoContent {
		b, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return err
	}

	return nil
}
