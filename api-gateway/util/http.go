package util

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// NewURLWithValues add values to base url
func NewURLWithValues(rawURL string, values map[string]string) (*url.URL, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return url, err
	}
	q := url.Query()
	for k, v := range values {
		q.Add(k, v)
	}

	url.RawQuery = q.Encode()
	return url, nil
}

// Http http请求
func Http(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	c := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// HttpWithTimeout http请求
func HttpWithTimeout(url string, timeout time.Duration) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
