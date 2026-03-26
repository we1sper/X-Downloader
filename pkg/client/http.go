package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/we1sper/X-Downloader/pkg/backoff"
)

type Option struct {
	Proxy                 string
	Retry                 uint64
	TimeoutInMilliseconds uint64
	BackoffStrategy       string
}

type HttpClient struct {
	client http.Client
	header http.Header

	option *Option
}

func NewHttpClient(option *Option) (*HttpClient, error) {
	httpClient := &HttpClient{
		client: http.Client{},
		header: http.Header{},
		option: option,
	}

	if len(option.BackoffStrategy) > 0 {
		// Pre-check backoff.
		if _, err := backoff.Get(option.BackoffStrategy, nil); err != nil {
			return nil, fmt.Errorf("pre-check backoff '%s' error: %v", option.BackoffStrategy, err)
		}
	} else {
		option.BackoffStrategy = "SimpleExponential"
	}

	// Configure http proxy.
	if len(option.Proxy) > 0 {
		if _proxy, err := url.Parse(option.Proxy); err != nil {
			return nil, fmt.Errorf("parse proxy '%s' error: %v", option.Proxy, err)
		} else {
			httpClient.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(_proxy),
			}
		}
	}

	if option.TimeoutInMilliseconds > 0 {
		httpClient.client.Timeout = time.Duration(option.TimeoutInMilliseconds) * time.Millisecond
	}

	return httpClient, nil
}

func (httpClient *HttpClient) Call(caller func() (*http.Response, error)) (resp *http.Response, err error) {
	_backoff, _ := backoff.Get(httpClient.option.BackoffStrategy, &backoff.Option{
		BaseInMilliseconds: 200,
	})

	for try := 0; try <= int(httpClient.option.Retry); try++ {
		if try > 0 {
			time.Sleep(_backoff.Next())
		}
		if resp, err = caller(); err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("http status code is %d", resp.StatusCode)
			// Terminate when meets '401', '403', '404' and '429'.
			if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden ||
				resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusTooManyRequests {
				return nil, err
			}
			continue
		}
		break
	}

	if err != nil {
		return nil, fmt.Errorf("call failed after %d tries, the last error: %v", httpClient.option.Retry+1, err)
	}

	return resp, nil
}

func (httpClient *HttpClient) AddHeader(key, value string) *HttpClient {
	httpClient.header.Add(key, value)
	return httpClient
}

func (httpClient *HttpClient) SetCookie(cookie string) *HttpClient {
	httpClient.AddHeader("Cookie", cookie)
	return httpClient
}
