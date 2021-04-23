package github

import (
	"bytes"
	"context"
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
	"github-release-admin/util"
)

type Client struct {
	ctx        context.Context
	baseURL    string
	baseHeader http.Header
	Header     http.Header
	Body       io.Reader
}

const GITHUB_API_URL = "https://api.github.com"

func getGitHubAPIURL() (string, error) {
	// use GITHUB_API_URL if define
	if s, found := util.Getenv("GITHUB_API_URL"); found && s != "" {
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("invalid GITHUB_API_URL environment variable: %w", err)
		} else if u.Path == "/" {
			u.Path = ""
		}

		if u.Scheme == "" || u.Host == "" || u.Path != "" ||
			u.RawQuery != "" || u.Fragment != "" {
			return "", fmt.Errorf("invalid GITHUB_API_URL environment variable: invalid url")
		}
		return u.String(), nil
	}

	return GITHUB_API_URL, nil
}

var ReOwnerName = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
var ReRepoName = regexp.MustCompile(`^[\w-][\w.-]*$`)

func New(ctx context.Context, repo string) (*Client, error) {
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

	// use GITHUB_API_URL if define
	baseURL, err := getGitHubAPIURL()
	if err != nil {
		return nil, err
	}

	c := &Client{
		ctx:     ctx,
		baseURL: baseURL + "/repos/" + repo,
		baseHeader: http.Header{
			"Accept": {"application/vnd.github.v3+json"},
		},
	}

	// use GITHUB_TOKEN if define
	if token, found := util.Getenv("GITHUB_TOKEN"); found && token != "" {
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

	req, err := http.NewRequestWithContext(c.ctx, method, url, body)
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

var ReMultipleSlashes = regexp.MustCompile("/+")

func resolveEndpoint(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint %w", err)
	}

	u = u.ResolveReference(&url.URL{})
	if u.User != nil || u.Scheme != "" || u.Host != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid endpoint %q", s)
	}

	u.Path = ReMultipleSlashes.ReplaceAllString(u.Path, "/")
	return u.String(), nil
}

func (c *Client) request(method, endpoint string) (*http.Response, error) {
	u, err := resolveEndpoint(endpoint)
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

func (c *Client) upload(method, endpoint string, body io.Reader, size int64, mime string) (*http.Response, error) {
	c.Body = body
	req, err := c.createRequest(method, endpoint)
	if err != nil {
		return nil, err
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", mime)
	req.Header.Set("Expect", "100-continue")

	if err = c.log(req, false); err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) PostUpload(endpoint string, body io.Reader, size int64, mime string) (*http.Response, error) {
	return c.upload("POST", endpoint, body, size, mime)
}

func (c *Client) DownloadAsset(id int, pathname string) error {
	dir := os.TempDir()
	f, err := os.CreateTemp(dir, "ghr-download-*")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	u, err := resolveEndpoint(fmt.Sprintf("/releases/assets/%d", id))
	if err != nil {
		return err
	}

	req, err := c.createRequest("GET", c.baseURL+u)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")

	if err = c.log(req, false); err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusOK:
		size := rsp.ContentLength
		if n, err := io.Copy(f, rsp.Body); err != nil {
			return err
		} else if n != size {
			return fmt.Errorf("unable to download the required file size %d/%d", n, size)
		} else if err = os.Rename(f.Name(), pathname); err != nil {
			return err
		}
		return nil

	case http.StatusNotFound:
		return nil

	default:
		b, err := httputil.DumpResponse(rsp, false)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return err
	}
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

func (c *Client) DeleteTag(tag string) error {
	rsp, err := c.Delete(fmt.Sprintf("/git/refs/tags/%s", tag))
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNoContent, http.StatusUnprocessableEntity:
		return nil

	default:
		b, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return err
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
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Label              string `json:"label"`
	ContentType        string `json:"content_type"`
	Size               int    `json:"size"`
	URL                string `json:"url"`
	BrowserDownloadURL string `json:"browser_download_url"`
	DownloadCount      int    `json:"download_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	Uploader           Author `json:"uploader"`
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
	UploadURL       string  `json:"upload_url,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	PublishedAt     string  `json:"published_at,omitempty"`
	Author          Author  `json:"author,omitempty"`
	Assets          []Asset `json:"assets,omitempty"`
}

var ReUploadURLSuffix = regexp.MustCompile("/assets[^/]*$")

func (r *Release) UploadAsset(c *Client, name string, body io.Reader, size int64, mime string) error {
	baseURL := ReUploadURLSuffix.ReplaceAllString(r.UploadURL, "")
	endpoint := fmt.Sprintf("%s/assets?name=%s", baseURL, name)
	rsp, err := c.upload("POST", endpoint, body, size, mime)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusCreated:
		return nil

	default:
		b, err := httputil.DumpResponse(rsp, true)
		if err == nil {
			err = fmt.Errorf("%s", b)
		}
		return err
	}
}

type ListRelease struct {
	NextPage int
	Releases []*Release
}

// ReLinkNext is used to check the Link header
// 	<https://api.github.com/repositories/194783954/releases?per_page=1&page=2>; rel=\"next\"
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

type FetchReleaseCallback func(v *Release, page int) error

func (c *Client) FetchRelease(page, itemsPerPage int, fn FetchReleaseCallback) error {
	if page < 1 {
		page = 1
	}
	if itemsPerPage < 1 {
		itemsPerPage = 20
	}

	for page > 0 {
		list, err := c.ListReleases(itemsPerPage, page)
		if err != nil {
			return err
		}
		for _, v := range list.Releases {
			if err = fn(v, page); err != nil {
				return err
			}
		}
		page = list.NextPage
	}

	return nil
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

func (c *Client) GetRelease(id int) (*Release, error) {
	rsp, err := c.Get(fmt.Sprintf("/releases/%d", id))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNotFound:
		return nil, nil

	case http.StatusOK:
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

func (c *Client) GetReleaseByTagName(tag string) (*Release, error) {
	rsp, err := c.Get(fmt.Sprintf("/releases/tags/%s", tag))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNotFound:
		return nil, nil

	case http.StatusOK:
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

func (c *Client) GetReleaseLatest() (*Release, error) {
	rsp, err := c.Get("/releases/latest")
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusNotFound:
		return nil, nil

	case http.StatusOK:
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
