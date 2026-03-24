package x

import (
	"fmt"

	"X-Downloader/pkg/client"
	"X-Downloader/pkg/util"
	"X-Downloader/x/api"
	"X-Downloader/x/config"
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

func LoadLatestMetadata(saveDir string, barrierCandidate uint64) (*Metadata, []string, error) {
	barriers := make([]string, 0)
	metadataFiles, err := util.FindJsonFiles(saveDir)
	if err != nil || len(metadataFiles) == 0 {
		return nil, barriers, err
	}
	latestMetadata, err := util.LoadFile[Metadata](metadataFiles[0].Path)
	if err != nil {
		return nil, barriers, err
	}
	// Use multiple barriers to prevent delta mode failure due to deleted tweets.
	for i := 0; i < int(barrierCandidate) && i < len(latestMetadata.Tweets); i++ {
		barriers = append(barriers, latestMetadata.Tweets[i].ID)
	}
	return latestMetadata, barriers, nil
}

func MergeTweets(previous, delta []*api.Tweet) []*api.Tweet {
	rough := make([]*api.Tweet, 0, len(previous)+len(delta))
	rough = append(rough, delta...)
	rough = append(rough, previous...)
	// Check duplicate tweets for robustness.
	merged := make([]*api.Tweet, 0)
	lookup := make(map[string]struct{})
	for _, t := range rough {
		if _, ok := lookup[t.ID]; !ok {
			merged = append(merged, t)
			lookup[t.ID] = struct{}{}
		}
	}
	return merged
}
