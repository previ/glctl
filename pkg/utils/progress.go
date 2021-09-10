package utils

import (
	"sync"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

type ProgressIndicator struct {
	progress      *mpb.Progress
	progressBars  map[string]*mpb.Bar
	progressMutex *sync.Mutex
}

func NewProgressIndicator() *ProgressIndicator {
	pi := &ProgressIndicator{}
	pi.progress = mpb.New(mpb.WithWidth(64))
	pi.progressBars = make(map[string]*mpb.Bar)
	pi.progressMutex = new(sync.Mutex)
	return pi
}
func (pi *ProgressIndicator) CreateBar(name string) {
	bar := pi.progress.AddBar(int64(100),
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name(name),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 60
				decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
			),
		),
	)
	pi.progressMutex.Lock()
	pi.progressBars[name] = bar
	pi.progressMutex.Unlock()
}

func (pi *ProgressIndicator) IncrementBar(name string) {
	pi.progressMutex.Lock()
	pi.progressBars[name].Increment()
	pi.progressMutex.Unlock()
}

func (pi *ProgressIndicator) CompleteBar(name string) {
	pi.progressMutex.Lock()
	pi.progressBars[name].SetTotal(0, true)
	pi.progressMutex.Unlock()
}
