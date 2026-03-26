package x

import (
	"fmt"
	"os"
	"testing"

	"github.com/we1sper/X-Downloader/x/config"
)

func prepare(configurer func(cfg *config.Config)) (*XClient, error) {
	cfg := config.NewConfig()
	cfg.Proxy = os.Getenv("PROXY")
	cfg.Cookie = os.Getenv("COOKIE")
	cfg.BearerToken = os.Getenv("BEARER_TOKEN")
	if configurer != nil {
		configurer(cfg)
	}
	return NewXClient(cfg)
}

func TestNewXClient(t *testing.T) {
	_, err := prepare(nil)
	if err != nil {
		t.Fatal("create x client error:", err)
	}
}

func TestXClient_QueryUserProfile(t *testing.T) {
	xClient, err := prepare(nil)
	if err != nil {
		t.Fatal("create x client error:", err)
	}
	profile, err := xClient.QueryProfile("elonmusk")
	if err != nil {
		t.Fatal("query user profile error:", err)
	}
	fmt.Println(*profile)
}

func TestXClient_QueryMediaTweets(t *testing.T) {
	xClient, err := prepare(nil)
	if err != nil {
		t.Fatal("create x client error:", err)
	}
	profile, err := xClient.QueryProfile("elonmusk")
	if err != nil {
		t.Fatal("query user profile error:", err)
	}
	result, err := xClient.QueryMediaTweets(profile.ID, "", nil)
	if err != nil {
		t.Fatal("query user media tweets error:", err)
	}
	fmt.Println(*result)
}

func TestXClient_QueryTweets(t *testing.T) {
	xClient, err := prepare(nil)
	if err != nil {
		t.Fatal("create x client error:", err)
	}
	profile, err := xClient.QueryProfile("elonmusk")
	if err != nil {
		t.Fatal("query user profile error:", err)
	}
	result, err := xClient.QueryTweets(profile.ID, "", nil)
	if err != nil {
		t.Fatal("query user tweets error:", err)
	}
	fmt.Println(*result)
}

func TestXClient_DownloadTweets(t *testing.T) {
	configurer := func(cfg *config.Config) {
		cfg.SaveDir = "./testdata"
	}
	xClient, err := prepare(configurer)
	if err != nil {
		t.Fatal("create x client error:", err)
	}
	_, err = xClient.DownloadTweets("elonmusk")
	if err != nil {
		t.Fatal("download tweets error:", err)
	}
}

func TestXClient_DownloadMediaTweets(t *testing.T) {
	configurer := func(cfg *config.Config) {
		cfg.SaveDir = "./testdata"
	}
	xClient, err := prepare(configurer)
	if err != nil {
		t.Fatal("create x client error:", err)
	}
	_, err = xClient.DownloadMediaTweets("elonmusk")
	if err != nil {
		t.Fatal("download media tweets error:", err)
	}
}
