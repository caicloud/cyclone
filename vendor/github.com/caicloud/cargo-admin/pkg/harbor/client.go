package harbor

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/caicloud/cargo-admin/pkg/errors"

	"github.com/caicloud/nirvana/log"
)

type Client struct {
	config   *Config
	baseURL  string
	client   *http.Client
	coockies []*http.Cookie
}

func NewClient(host, username, password string) (*Client, error) {
	return newClient(&Config{host, username, password})
}

func newClient(conf *Config) (*Client, error) {
	baseURL := strings.TrimRight(conf.Host, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	cookies, err := LoginAndGetCookies(conf)
	if err != nil {
		log.Errorf("login harbor: %s error: %v during background", conf.Host, err)
		return nil, err
	}
	log.Infof("harbor %s cookies has been refreshed", conf.Host)

	return &Client{
		config:   conf,
		baseURL:  baseURL,
		client:   http.DefaultClient,
		coockies: cookies,
	}, nil
}

// do creates request and authorizes it if authorizer is not nil
func (c *Client) do(method, relativePath string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + relativePath
	log.Infof("%s %s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i, _ := range c.coockies {
		req.AddCookie(c.coockies[i])
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return nil, err
	}
	if resp.StatusCode/100 == 5 || resp.StatusCode == 401 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, ErrorUnknownInternal.Error(err)
		}
		defer resp.Body.Close()

		log.Errorf("unexpected %d error from harbor: %s", resp.StatusCode, b)
		log.Errorf("need to refresh harbor: %s 's cookies now! refreshCookies error: %v", c.config.Host, c.refreshCookies())
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("harbor internal error: %s", b))
	}
	return resp, nil
}

func (c *Client) refreshCookies() error {
	cookies, err := LoginAndGetCookies(c.config)
	if err != nil {
		log.Errorf("refresh harbor: %s 's cookies error: %v", c.config.Host, err)
		return err
	}
	c.coockies = cookies
	return nil
}

func (c *Client) GetConfig() *Config {
	return c.config
}
