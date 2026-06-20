package morse

import (
	"time"
)

func WPMToElementDuration(wpm uint) time.Duration {
	ditDuration := 1200 / wpm
	return time.Duration(ditDuration) * time.Millisecond
}

func ElementDurationToWPM(duration time.Duration) uint {
	ditDuration := duration.Milliseconds()
	if ditDuration == 0 {
		return 0
	}
	wpm := 1200 / uint(ditDuration)
	return wpm
}

func DitDahsToIntervals(ditDahs string, wpm uint) []time.Duration {
	var intervals []time.Duration
	ditDuration := WPMToElementDuration(wpm)
	for _, c := range ditDahs {
		switch c {
		case '.':
			intervals = append(intervals, ditDuration)
		case '-':
			intervals = append(intervals, ditDuration*3)
		}
		intervals = append(intervals, -ditDuration)
	}
	return intervals
}
