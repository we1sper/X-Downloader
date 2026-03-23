package x

import (
	"fmt"
	"os"
	"testing"

	"X-Downloader/pkg/config"
)

func prepare() (*XClient, error) {
	cfg := config.NewConfig()
	cfg.Proxy = os.Getenv("PROXY")
	cfg.Cookie = os.Getenv("COOKIE")
	cfg.BearerToken = os.Getenv("BEARER_TOKEN")
	return NewXClient(cfg)
}

func TestNewXClient(t *testing.T) {
	_, err := prepare()
	if err != nil {
		t.Fatal("create x client error:", err)
	}
}

func TestXClient_QueryUserProfile(t *testing.T) {
	xClient, err := prepare()
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
	xClient, err := prepare()
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
	xClient, err := prepare()
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
