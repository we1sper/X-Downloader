package x

import (
	"sync"
	"time"

	"X-Downloader/pkg/log"
	"X-Downloader/x/api"
)

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

type TaskMetadata struct {
	expected      uint64
	count         uint64
	succeed       uint64
	failed        uint64
	skipped       uint64
	finished      bool
	start         time.Time
	end           time.Time
	shutdownHooks []func()
	lock          sync.RWMutex
}

func (metadata *TaskMetadata) RegisterShutdownHook(hook func()) *TaskMetadata {
	metadata.shutdownHooks = append(metadata.shutdownHooks, hook)
	return metadata
}

func (metadata *TaskMetadata) update(updater func(), afterUpdateHook func(progress float64), finalizer func()) {
	metadata.lock.Lock()
	defer metadata.lock.Unlock()

	if !metadata.finished {
		updater()
		metadata.count++
		if metadata.count >= metadata.expected {
			metadata.finished = true
			metadata.end = time.Now()
		}
	}

	progress := float64(metadata.count) / float64(metadata.expected) * 100.0
	afterUpdateHook(progress)

	if metadata.finished {
		finalizer()
		for _, hook := range metadata.shutdownHooks {
			hook()
		}
	}
}

type DownloadTask struct {
	*TaskMetadata
	*api.Media
	UserName string
}

func (task *DownloadTask) ReportSucceed(elapsed time.Duration) {
	updater := func() {
		task.succeed++
	}
	afterUpdateHook := func(progress float64) {
		log.Infof("[Downloader][%s][%.2f%%][%d/%d][%s] '%s' => succeed using %.2fs",
			task.UserName, progress, task.count, task.expected, task.Type, task.Name, elapsed.Seconds())
	}
	task.update(updater, afterUpdateHook, task.finalize)
}

func (task *DownloadTask) ReportFailed(elapsed time.Duration, err error) {
	updater := func() {
		task.failed++
	}
	afterUpdateHook := func(progress float64) {
		log.Errorf("[Downloader][%s][%.2f%%][%d/%d][%s] '%s' => failed after %.2f, caused by: %v",
			task.UserName, progress, task.count, task.expected, task.Type, task.Name, elapsed.Seconds(), err)
	}
	task.update(updater, afterUpdateHook, task.finalize)
}

func (task *DownloadTask) ReportSkipped() {
	updater := func() {
		task.succeed++
		task.skipped++
	}
	afterUpdateHook := func(progress float64) {
		log.Infof("[Downloader][%s][%.2f%%][%d/%d][%s] '%s' => skipped",
			task.UserName, progress, task.count, task.expected, task.Type, task.Name)
	}
	task.update(updater, afterUpdateHook, task.finalize)
}

func (task *DownloadTask) finalize() {
	elapsed := task.end.Sub(task.start)
	if seconds := elapsed.Seconds(); seconds < 60 {
		log.Infof("[Downloader][%s] finished using %.2fs: new=%d, skipped=%d, failed=%d",
			task.UserName, seconds, task.succeed-task.skipped, task.skipped, task.failed)
	} else {
		log.Infof("[Downloader][%s] finished using %.2fm: new=%d, skipped=%d, failed=%d",
			task.UserName, elapsed.Minutes(), task.succeed-task.skipped, task.skipped, task.failed)
	}
}
