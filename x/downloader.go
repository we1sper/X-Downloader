package x

import (
	"sync"
	"time"

	"github.com/we1sper/X-Downloader/pkg/log"
	"github.com/we1sper/X-Downloader/x/api"
)

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

func (x *XClient) downloader(id int) {
	log.Infof("[Downloader][%d] starts", id)

	defer func() {
		log.Infof("[Downloader][%d] stops", id)
		x.wg.Done()
	}()

	for {
		select {
		case task := <-x.taskChan:
			if task != nil {
				start := time.Now()
				skipped, err := x.Download(task)
				elapsed := time.Since(start)
				if err != nil {
					task.ReportFailed(elapsed, err)
				} else if skipped {
					task.ReportSkipped()
				} else {
					task.ReportSucceed(elapsed)
				}
			}
		case <-x.ctx.Done():
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
