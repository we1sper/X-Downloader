package x

import (
	"fmt"
	"math/rand"
	"path"
	"time"

	"github.com/we1sper/X-Downloader/pkg/log"
	"github.com/we1sper/X-Downloader/x/api"
)

func (x *XClient) scroller(screenName string, caller func(userID, cursor string, barriers []string) (*QueryResult, error), all bool) (*Metadata, error) {
	start := time.Now()

	log.Infof("[Analyzer][%s] starts", screenName)
	defer func() {
		elapsed := time.Since(start)
		log.Infof("[Analyzer][%s] finished using %.2fs (%.2fm)", screenName, elapsed.Seconds(), elapsed.Minutes())
	}()

	// First, query user profile as we need userID for further queries.
	profile, err := x.QueryProfile(screenName)
	if err != nil {
		return nil, fmt.Errorf("failed to query profile: %v", err)
	}

	var previous *Metadata
	barriers := make([]string, 0)
	// Use 'screenName' as the directory name.
	saveDir := path.Join(x.cfg.SaveDir, screenName)

	if x.cfg.Delta {
		previous, barriers, err = LoadLatestMetadata(saveDir, x.cfg.BarrierCandidate)
		if err != nil {
			log.Errorf("[Analyzer][%s] delta mode enabled, but failed to load the latest metadata: %v => fallback to standard mode", screenName, err)
		} else {
			log.Infof("[Analyzer][%s] delta mode enabled", screenName)
		}
	}

	step := 0
	cursor := ""
	earlyStopped := false
	newly := &Metadata{
		UserProfile: profile,
		Tweets:      make([]*api.Tweet, 0),
	}

	overall := profile.MediaCount
	if all {
		overall = profile.ListedCount
	}

	// Scroll.
	for {
		queryResult, err := caller(profile.ID, cursor, barriers)
		if err != nil {
			log.Errorf("[Analyzer][%s] query with cursor '%s' error: %v", screenName, cursor, err)
		}

		step += len(queryResult.Tweets)
		cursor = queryResult.NextCursor
		earlyStopped = queryResult.EarlyStopped
		newly.Tweets = append(newly.Tweets, queryResult.Tweets...)

		if overall != 0 {
			progress := float64(step) / overall * 100.0
			log.Infof("[Analyzer][%s] %.2f%%", screenName, progress)
		}

		if all {
			// The 'UserTweets' API can easily hit rate limits.
			cooling := time.Duration(rand.Intn(10)) * time.Second
			log.Infof("[Analyzer][%s] cooling %vs", screenName, cooling.Seconds())
			time.Sleep(cooling)
		}

		if earlyStopped {
			log.Infof("[Analyzer][%s] touch barrier => early stopped", screenName)
			break
		}

		if len(queryResult.Tweets) == 0 {
			break
		}
	}

	if earlyStopped && previous != nil {
		// Merge is needed for delta mode.
		newly.Tweets = MergeTweets(previous.Tweets, newly.Tweets)
	}

	if err := x.saveMetadata(saveDir, newly); err != nil {
		log.Errorf("[Analyzer][%s] failed to save metadata: %v => continue subsequent procedures", screenName, err)
	} else {
		log.Infof("[Analyzer][%s] metadata saved", screenName)
	}

	if x.cfg.Download {
		log.Infof("[Analyzer][%s] dispatch tasks", screenName)
		taskMetadata := &TaskMetadata{
			start: time.Now(),
		}
		taskMetadata.RegisterShutdownHook(x.Stop)
		// Count download task number.
		for _, tweet := range newly.Tweets {
			taskMetadata.expected += uint64(len(tweet.Media))
		}
		// Dispatch download tasks.
		for idx := len(newly.Tweets) - 1; idx >= 0; idx-- {
			for _, m := range newly.Tweets[idx].Media {
				x.taskChan <- &DownloadTask{
					TaskMetadata: taskMetadata,
					Media:        m,
					UserName:     screenName,
				}
			}
		}
		log.Infof("[Analyzer][%s] dispatch finished", screenName)
	} else {
		x.Stop()
	}

	return newly, nil
}
