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
