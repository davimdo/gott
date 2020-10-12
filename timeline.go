package gott

import "time"

type Timeline struct {
	Start time.Time
	End   time.Time
}

func (tl *Timeline) SeekByPercentage(percentage int) time.Time {
	seekDuration := time.Duration((int(tl.Duration()) * percentage) / 100)
	return tl.Start.Add(seekDuration)
}

func (tl *Timeline) Duration() time.Duration {
	return time.Since(tl.Start) - time.Since(tl.End)
}
