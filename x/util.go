package x

import (
	"fmt"

	"X-Downloader/pkg/client"
	"X-Downloader/pkg/config"
	"X-Downloader/pkg/util"
)

func createBaseHttpClient(cfg *config.Config) (*client.HttpClient, error) {
	httpClient, err := client.NewHttpClient(&client.Option{
		Proxy:                 cfg.Proxy,
		Retry:                 cfg.Retry,
		TimeoutInMilliseconds: cfg.Timeout,
	})
	if err != nil {
		return nil, err
	}

	httpClient.AddHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	return httpClient, nil
}

func createHttpClient(cfg *config.Config) (*client.HttpClient, error) {
	httpClient, err := createBaseHttpClient(cfg)
	if err != nil {
		return nil, err
	}

	httpClient.SetCookie(cfg.Cookie).
		AddHeader("X-Csrf-Token", util.ExtractFromCookie(cfg.Cookie, "ct0")).
		AddHeader("Authorization", fmt.Sprintf("Bearer %s", cfg.BearerToken))

	return httpClient, nil
}
