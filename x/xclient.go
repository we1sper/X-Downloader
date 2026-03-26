package x

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/we1sper/X-Downloader/pkg/log"
	"github.com/we1sper/X-Downloader/pkg/util"

	"github.com/we1sper/X-Downloader/pkg/client"
	"github.com/we1sper/X-Downloader/pkg/value"
	"github.com/we1sper/X-Downloader/x/api"
	"github.com/we1sper/X-Downloader/x/config"
)

type Metadata struct {
	UserProfile *api.UserProfile `json:"user_profile"`
	Tweets      []*api.Tweet     `json:"tweets"`
}

type QueryResult struct {
	NextCursor     string
	EarlyStopped   bool
	BarrierTouched string
	Tweets         []*api.Tweet
}

func (result *QueryResult) AppendTweet(tweet *api.Tweet, barriers []string) bool {
	for _, barrier := range barriers {
		if barrier == tweet.ID {
			result.EarlyStopped = true
			result.BarrierTouched = barrier
			break
		}
	}
	if !result.EarlyStopped {
		result.Tweets = append(result.Tweets, tweet)
	}
	return !result.EarlyStopped
}

type XClient struct {
	queryClient    *client.HttpClient
	downloadClient *client.HttpClient
	cfg            *config.Config
	wg             *sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	taskChan       chan *DownloadTask
}

func NewXClient(cfg *config.Config) (*XClient, error) {
	xClient := &XClient{
		cfg:      cfg,
		wg:       &sync.WaitGroup{},
		taskChan: make(chan *DownloadTask, 16384),
	}

	if queryClient, err := createHttpClient(cfg); err != nil {
		return nil, fmt.Errorf("create query client error: %v", err)
	} else {
		xClient.queryClient = queryClient
	}

	if downloadClient, err := createBaseHttpClient(cfg); err != nil {
		return nil, fmt.Errorf("create download client error: %v", err)
	} else {
		xClient.downloadClient = downloadClient
	}

	xClient.ctx, xClient.cancel = context.WithCancel(context.Background())

	return xClient, nil
}

func (x *XClient) Start() {
	for id := 0; id < int(x.cfg.Downloader); id++ {
		x.wg.Add(1)
		go x.downloader(id)
	}
}

func (x *XClient) Stop() {
	x.cancel()

	time.Sleep(1 * time.Second)

	drain := func() {
		for {
			select {
			case _ = <-x.taskChan:
			default:
				return
			}
		}
	}

	drain()

	close(x.taskChan)
}

func (x *XClient) Hang() {
	x.wg.Wait()
}

func (x *XClient) Download(task *DownloadTask) (bool, error) {
	filePath := path.Join(x.cfg.SaveDir, task.UserName, task.Name)
	if !x.cfg.Overwrite && util.IsFileExist(filePath) {
		return true, nil
	}

	resp, err := x.downloadClient.Get(task.DownloadURL)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err = util.SaveFromStream(filePath, resp.Body); err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("download task {'type': '%s', 'name': '%s', 'url': '%s'} failed: http status code is %d",
			task.Type, task.Name, task.DownloadURL, resp.StatusCode)
	}

	return false, nil
}

func (x *XClient) QueryProfile(screenName string) (*api.UserProfile, error) {
	query, err := api.NewUserProfileQuery(screenName)
	if err != nil {
		return nil, fmt.Errorf("create user profile query error: %v", err)
	}

	resp, err := x.queryClient.Get(query.EncodedUrl)
	if err != nil {
		return nil, fmt.Errorf("execute user profile query error: %v", err)
	}

	defer resp.Body.Close()

	v, err := value.NewValueFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("create value from response body error: %v", err)
	}

	r := v.Get("data", "user", "result")
	l := r.Get("legacy")

	profile := &api.UserProfile{
		ID:                r.Get("rest_id").Unsafe().String(),
		Name:              l.Get("name").Unsafe().String(),
		ScreenName:        l.Get("screen_name").Unsafe().String(),
		CreatedAt:         l.Get("created_at").Unsafe().String(),
		Description:       l.Get("description").Unsafe().String(),
		FollowersCount:    l.Get("followers_count").Unsafe().Float64(),
		FriendsCount:      l.Get("friends_count").Unsafe().Float64(),
		MediaCount:        l.Get("media_count").Unsafe().Float64(),
		ListedCount:       l.Get("listed_count").Unsafe().Float64(),
		Location:          l.Get("location").Unsafe().String(),
		PinnedTweetIds:    l.Get("pinned_tweet_ids_str").Unsafe().StringSlice(),
		PossiblySensitive: l.Get("possibly_sensitive").Unsafe().Bool(),
		FetchTime:         time.Now().Format(time.DateTime),
	}

	// Extract birthday.
	b := r.Get("legacy_extended_profile", "birthdate")
	if b.Success() {
		birthday := ""
		if b.Get("visibility").Unsafe().String() == "Public" {
			day := b.Get("day").Unsafe().Float64()
			month := b.Get("month").Unsafe().Float64()
			birthday = fmt.Sprintf("%v-%v", month, day)
			if b.Get("year_visibility").Unsafe().String() == "Public" {
				if year := b.Get("year").Unsafe().String(); len(year) > 0 {
					birthday = birthday + "-" + year
				}
			}
		}
		profile.Birthday = birthday
	}

	if r.Get("verification_info", "reason", "description", "text").Success() {
		profile.IsVerified = true
	}

	banner := l.Get("profile_banner_url").Unsafe().String()
	if len(banner) > 0 {
		profile.BannerURL = banner
		profile.DownloadBannerURL = banner + "/1500x500"
	}

	image := l.Get("profile_image_url_https").Unsafe().String()
	if len(image) > 0 {
		profile.ImageURL = image
		profile.DownloadImageURL = strings.Replace(image, "_normal.jpg", "_400x400.jpg", -1)
	}

	return profile, nil
}

func (x *XClient) QueryMediaTweets(userID, cursor string, barriers []string) (*QueryResult, error) {
	queryResult, v, err := x.prepare(func() string {
		query, _ := api.NewUserMediaQuery(userID, cursor)
		return query.EncodedUrl
	})
	if err != nil {
		return queryResult, err
	}

	instructions := v.Get("data", "user", "result", "timeline_v2", "timeline", "instructions")

	resultProcessor := func(result *value.Value, pinned bool) bool {
		tweet := x.handleResult(result)
		tweet.Pinned = pinned
		return queryResult.AppendTweet(tweet, barriers)
	}

	itemProcessor := func(item *value.Value) bool {
		itemContent := item.Get("item", "itemContent")
		itemType := itemContent.Get("itemType").Unsafe().String()

		if itemType == "TimelineTweet" {
			pinned := itemContent.Get("socialContext", "contextType").Unsafe().String() == "Pin"
			result := itemContent.Get("tweet_results", "result")
			resultType := result.Get("__typename").Unsafe().String()

			switch resultType {
			case "Tweet":
				return resultProcessor(result, pinned)
			case "TweetWithVisibilityResults":
				return resultProcessor(result.Get("tweet"), pinned)
			}
		}

		return true
	}

	entryProcessor := func(entry *value.Value) bool {
		entryType := entry.Get("content", "entryType").Unsafe().String()

		switch entryType {
		case "TimelineTimelineCursor":
			x.cursorProcessor(entry, func(cursor string) {
				queryResult.NextCursor = cursor
			})
		case "TimelineTimelineModule":
			for _, item := range entry.Get("content", "items").Unsafe().ValueSlice() {
				if !itemProcessor(item) {
					return false
				}
			}
		}

		return true
	}

	instructionProcessor := func(instruction *value.Value) bool {
		instructionType := instruction.Get("type").Unsafe().String()

		switch instructionType {
		case "TimelineAddEntries":
			for _, entry := range instruction.Get("entries").Unsafe().ValueSlice() {
				if !entryProcessor(entry) {
					return false
				}
			}
		case "TimelineAddToModule":
			for _, item := range instruction.Get("moduleItems").Unsafe().ValueSlice() {
				if !itemProcessor(item) {
					return false
				}
			}
		}

		return true
	}

	for _, instruction := range instructions.Unsafe().ValueSlice() {
		if !instructionProcessor(instruction) {
			break
		}
	}

	return queryResult, nil
}

func (x *XClient) QueryTweets(userID, cursor string, barriers []string) (*QueryResult, error) {
	queryResult, v, err := x.prepare(func() string {
		query, _ := api.NewUserTweetsQuery(userID, cursor)
		return query.EncodedUrl
	})
	if err != nil {
		return queryResult, err
	}

	instructions := v.Get("data", "user", "result", "timeline", "timeline", "instructions")

	processEntry := func(entry *value.Value) bool {
		itemContent := entry.Get("content", "itemContent")
		tweet := x.handleResult(itemContent.Get("tweet_results", "result"))
		tweet.Pinned = itemContent.Get("socialContext", "contextType").Unsafe().String() == "Pin"
		return queryResult.AppendTweet(tweet, barriers)
	}

	entryProcessor := func(entry *value.Value) bool {
		entryType := entry.Get("content", "entryType").Unsafe().String()

		switch entryType {
		case "TimelineTimelineItem":
			return processEntry(entry)
		case "TimelineTimelineCursor":
			x.cursorProcessor(entry, func(cursor string) {
				queryResult.NextCursor = cursor
			})
		}

		return true
	}

	instructionProcessor := func(instruction *value.Value) bool {
		instructionType := instruction.Get("type").Unsafe().String()

		switch instructionType {
		case "TimelinePinEntry":
			return entryProcessor(instruction.Get("entry"))
		case "TimelineAddEntries":
			for _, entry := range instruction.Get("entries").Unsafe().ValueSlice() {
				if !entryProcessor(entry) {
					return false
				}
			}
		}

		return true
	}

	for _, instruction := range instructions.Unsafe().ValueSlice() {
		if !instructionProcessor(instruction) {
			break
		}
	}

	return queryResult, nil
}

func (x *XClient) DownloadMediaTweets(screenName string) (*Metadata, error) {
	return x.scroller(screenName, x.QueryMediaTweets, false)
}

func (x *XClient) DownloadTweets(screenName string) (*Metadata, error) {
	return x.scroller(screenName, x.QueryTweets, true)
}

func (x *XClient) prepare(urlProvider func() string) (*QueryResult, *value.Value, error) {
	queryResult := &QueryResult{
		NextCursor:     "",
		EarlyStopped:   false,
		BarrierTouched: "",
		Tweets:         make([]*api.Tweet, 0),
	}

	url := urlProvider()

	resp, err := x.queryClient.Get(url)
	if err != nil {
		return queryResult, nil, err
	}

	defer resp.Body.Close()

	v, err := value.NewValueFromReader(resp.Body)
	if err != nil {
		return queryResult, nil, err
	}

	return queryResult, v, nil
}

func (x *XClient) handleResult(r *value.Value) *api.Tweet {
	l := r.Get("legacy")

	tweet := &api.Tweet{
		ID:            r.Get("rest_id").Unsafe().String(),
		Text:          l.Get("full_text").Unsafe().String(),
		CreatedAt:     l.Get("created_at").Unsafe().String(),
		QuoteCount:    l.Get("quote_count").Unsafe().Float64(),
		ReplyCount:    l.Get("reply_count").Unsafe().Float64(),
		RetweetCount:  l.Get("retweet_count").Unsafe().Float64(),
		BookmarkCount: l.Get("bookmark_count").Unsafe().Float64(),
		FavoriteCount: l.Get("favorite_count").Unsafe().Float64(),
		Retweeted:     l.Get("retweeted").Unsafe().Bool(),
		IsQuote:       l.Get("is_quote_status").Unsafe().Bool(),
		Quoted:        nil,
		Media:         make([]*api.Media, 0),
	}

	tweet.Link = x.extractTweetLink(tweet.Text)

	// Format 'CreatedAt'.
	formatedTime, _ := time.Parse(time.RubyDate, tweet.CreatedAt)
	tweet.CreatedAt = formatedTime.Local().Format(time.DateTime)

	if tweet.IsQuote {
		// We can get expanded text of quoted tweets if they are present.
		n := r.Get("note_tweet")
		if n.Get("is_expandable").Unsafe().Bool() {
			tweet.ExpandedText = n.Get("note_tweet_results", "result", "text").Unsafe().String()
		}
		// Process quoted permalink.
		permalink := l.Get("quoted_status_permalink", "url")
		if permalink.Success() {
			tweet.Link = permalink.Unsafe().String()
		}
	}

	for _, m := range l.Get("entities", "media").Unsafe().ValueSlice() {
		x.mediaProcessor(m, func(media *api.Media) {
			tweet.Media = append(tweet.Media, media)
		})
	}

	retweet := l.Get("retweeted_status_result", "result")
	if retweet.Success() {
		// SelfRetweeted condition: 'Retweeted' is false and 'retweeted_status_result' is present.
		tweet.SelfRetweeted = !tweet.Retweeted
		// Handle retweets recursively.
		tweet.Retweet = x.handleResult(retweet)
	}

	if tweet.IsQuote {
		quote := r.Get("quoted_status_result", "result")
		if quote.Success() {
			// Handle quoted tweets recursively.
			tweet.Quoted = x.handleResult(quote)
		}
	}

	return tweet
}

func (x *XClient) mediaProcessor(m *value.Value, acceptor func(media *api.Media)) {
	mediaType := m.Get("type").Unsafe().String()

	var url string

	if mediaType == "photo" {
		url = m.Get("media_url_https").Unsafe().String()
	} else if mediaType == "video" || mediaType == "animated_gif" {
		// The maximum resolution video is at the tail of 'variants'.
		url = m.Get("video_info", "variants").Tail().Get("url").Unsafe().String()
	} else {
		log.Warnf("[Analyzer] ignore unsupported media type '%s'", mediaType)
	}

	media := &api.Media{
		Type:        mediaType,
		Name:        util.ExtractFileNameFromURL(url),
		URL:         url,
		DownloadURL: url,
		Width:       m.Get("original_info", "width").Unsafe().Float64(),
		Height:      m.Get("original_info", "height").Unsafe().Float64(),
	}

	if mediaType == "photo" {
		// Append query parameter 'name=4096x4096' to download the largest photo?
		media.DownloadURL = media.DownloadURL + "?format=jpg&name=4096x4096"
	}

	acceptor(media)
}

func (x *XClient) cursorProcessor(entry *value.Value, acceptor func(cursor string)) {
	cursorType := entry.Get("content", "cursorType").Unsafe().String()

	// Get the bottom cursor as we follow the top-to-down scrolling strategy.
	if cursorType == "Bottom" {
		acceptor(entry.Get("content", "value").Unsafe().String())
	}
}

func (x *XClient) extractTweetLink(fullText string) string {
	candidate := fullText
	segments := strings.Split(fullText, " ")
	if len(segments) > 1 {
		candidate = segments[len(segments)-1]
	}
	if strings.HasPrefix(candidate, "https://") {
		return candidate
	}
	return ""
}

func (x *XClient) saveMetadata(saveDir string, metadata *Metadata) error {
	if err := util.CreateDirectory(saveDir); err != nil {
		return fmt.Errorf("create directory error: %v", err)
	}
	year, month, day := time.Now().Date()
	fileName := fmt.Sprintf("%v-%v-%v.json", year, int(month), day)
	filePath := path.Join(saveDir, fileName)
	if err := util.SaveToJsonFile(filePath, metadata); err != nil {
		return fmt.Errorf("save to file error: %v", err)
	}
	return nil
}
