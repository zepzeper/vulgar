package cron

import (
	"github.com/robfig/cron/v3"
)

var globalScheduler *scheduler

func getScheduler() *scheduler {
	if globalScheduler == nil {
		// Use standard parser that supports both 5-field and 6-field (with seconds) expressions
		parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		globalScheduler = &scheduler{
			c:    cron.New(cron.WithParser(parser)),
			jobs: make(map[cron.EntryID]*jobHandle),
		}
	}
	return globalScheduler
}
