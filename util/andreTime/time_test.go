package andretime

import (
	"testing"
	"time"

	assert2 "github.com/stretchr/testify/assert"
)

func GetTimeNow(andreeTime AndreTime) time.Time {
	return andreeTime.Now()
}

func Test_AndreeTime(t *testing.T) {

	t.Run("when Using fakeTime", func(t *testing.T) {
		timeNow := time.Date(2020, time.August, 31, 0, 0, 0, 0, time.UTC)
		fakeTime := NewFakeTimeAt(timeNow)
		response := GetTimeNow(fakeTime)

		assert2.Equal(t, fakeTime.Now(), response)
	})

	t.Run("when Using realTime", func(t *testing.T) {
		andreeTime := NewRealTime()
		response := GetTimeNow(andreeTime)

		assert2.NotEqual(t, andreeTime.Now(), response)
	})
}
