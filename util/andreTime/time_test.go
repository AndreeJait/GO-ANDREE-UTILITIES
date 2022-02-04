package tikettime

import (
	assert2 "github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func GetTimeNow (tiketTime TiketTime) time.Time {
	return tiketTime.Now()
}

func Test_TiketTime(t *testing.T) {

	t.Run("when Using fakeTime", func(t *testing.T) {
		timeNow := time.Date(2020, time.August, 31, 0, 0, 0, 0, time.UTC)
		fakeTime := NewFakeTimeAt(timeNow)
		response := GetTimeNow(fakeTime)

		assert2.Equal(t, fakeTime.Now(), response)
	})

	t.Run("when Using realTime", func(t *testing.T) {
		tiketTime := NewRealTime()
		response := GetTimeNow(tiketTime)

		assert2.NotEqual(t, tiketTime.Now(), response)
	})
}
