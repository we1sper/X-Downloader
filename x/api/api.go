package api

var XAPIs = map[string]string{}

type UserProfile struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	ScreenName        string   `json:"screen_name"`
	Birthday          string   `json:"birthday"`
	CreatedAt         string   `json:"created_at"`
	Description       string   `json:"description"`
	FollowersCount    float64  `json:"followers_count"`
	FriendsCount      float64  `json:"friends_count"`
	ListedCount       float64  `json:"listed_count"`
	MediaCount        float64  `json:"media_count"`
	Location          string   `json:"location"`
	PinnedTweetIds    []string `json:"pinned_tweet_ids"`
	PossiblySensitive bool     `json:"possibly_sensitive"`
	BannerURL         string   `json:"banner_url"`
	DownloadBannerURL string   `json:"download_banner_url"`
	ImageURL          string   `json:"image_url"`
	DownloadImageURL  string   `json:"download_image_url"`
	IsVerified        bool     `json:"is_verified"`
	FetchTime         string   `json:"fetch_time"`
}

type Tweet struct {
	ID            string   `json:"id"`
	Text          string   `json:"text"`
	Link          string   `json:"link"`
	CreatedAt     string   `json:"created_at"`
	QuoteCount    float64  `json:"quote_count"`
	ReplyCount    float64  `json:"reply_count"`
	RetweetCount  float64  `json:"retweet_count"`
	BookmarkCount float64  `json:"bookmark_count"`
	FavoriteCount float64  `json:"favorite_count"`
	Retweeted     bool     `json:"retweeted"`
	SelfRetweeted bool     `json:"self_retweeted"`
	IsQuote       bool     `json:"is_quote"`
	ExpandedText  string   `json:"expanded_text,omitempty"`
	Pinned        bool     `json:"pinned"`
	Retweet       *Tweet   `json:"retweet,omitempty"`
	Quoted        *Tweet   `json:"quoted,omitempty"`
	Media         []*Media `json:"media,omitempty"`
}

type Media struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	DownloadURL string  `json:"download_url"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
}
