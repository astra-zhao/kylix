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

// Put sends an HTTP PUT request with a string body.
func (c *THttpClient) Put(path, contentType, body string) (string, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("HttpClient.Put: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Put: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Put (read): %w", err)
	}
	return string(respBody), nil
}

// Delete sends an HTTP DELETE request and returns the response body.
func (c *THttpClient) Delete(path string) (string, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Delete: %w", err)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Delete: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpClient.Delete (read): %w", err)
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

// HttpPostJSON sends a POST request with a JSON body and decodes the JSON
// response into a map.
func HttpPostJSON(url, body string) (map[string]interface{}, error) {
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("HttpPostJSON: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HttpPostJSON (read): %w", err)
	}
	return JsonDecodeMap(string(respBody))
}

// HttpPut sends a PUT request with a body and returns the response body.
func HttpPut(url, contentType, body string) (string, error) {
	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("HttpPut: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HttpPut: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpPut (read): %w", err)
	}
	return string(respBody), nil
}

// HttpDelete sends a DELETE request and returns the response body.
func HttpDelete(url string) (string, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", fmt.Errorf("HttpDelete: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HttpDelete: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HttpDelete (read): %w", err)
	}
	return string(respBody), nil
}

// THttpResponse captures both the status code and body of an HTTP response.
type THttpResponse struct {
	Status int
	Body   string
}

// HttpDoGet performs a GET request and returns a THttpResponse with status
// code and body. Use this when you need to inspect the status code.
func HttpDoGet(url string) (*THttpResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HttpDoGet: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HttpDoGet (read): %w", err)
	}
	return &THttpResponse{Status: resp.StatusCode, Body: string(body)}, nil
}

// HttpDoPost performs a POST request and returns a THttpResponse.
func HttpDoPost(url, contentType, body string) (*THttpResponse, error) {
	resp, err := http.Post(url, contentType, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("HttpDoPost: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HttpDoPost (read): %w", err)
	}
	return &THttpResponse{Status: resp.StatusCode, Body: string(respBody)}, nil
}
