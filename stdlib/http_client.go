package stdlib

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// THttpClient is a simple HTTP client for Kylix programs.
type THttpClient struct {
	BaseURL    string
	Timeout    int // milliseconds, 0 = default (10s)
	Headers    map[string]string
	client     *http.Client
}

// NewHttpClient creates a new HTTP client.
func NewHttpClient(baseURL string, timeoutMs int) *THttpClient {
	if timeoutMs <= 0 {
		timeoutMs = 10000
	}
	return &THttpClient{
		BaseURL: baseURL,
		Timeout: timeoutMs,
		Headers: make(map[string]string),
		client:  &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

// SetHeader adds a request header to all subsequent requests.
func (c *THttpClient) SetHeader(key, value string) {
	c.Headers[key] = value
}

// Get sends an HTTP GET request and returns the response body as a string.
func (c *THttpClient) Get(path string) (string, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Get: %w", err)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Get (read): %w", err)
	}
	return string(body), nil
}

// Post sends an HTTP POST request with a string body.
func (c *THttpClient) Post(path, contentType, body string) (string, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("HttpClient.Post: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Post: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Post (read): %w", err)
	}
	return string(respBody), nil
}

// StatusCode sends a GET request and returns only the HTTP status code.
func (c *THttpClient) StatusCode(path string) (int, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// One-shot convenience functions (no client object needed)

// HttpGet fetches a URL and returns the body.
func HttpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HttpGet: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpGet (read): %w", err)
	}
	return string(body), nil
}

// HttpPost sends a POST request with body and returns the response.
func HttpPost(url, contentType, body string) (string, error) {
	resp, err := http.Post(url, contentType, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("HttpPost: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpPost (read): %w", err)
	}
	return string(respBody), nil
}

// HttpGetJSON fetches a URL, expecting JSON, and decodes it.
func HttpGetJSON(url string) (map[string]interface{}, error) {
	body, err := HttpGet(url)
	if err != nil {
		return nil, err
	}
	return JsonDecodeMap(body)
}
