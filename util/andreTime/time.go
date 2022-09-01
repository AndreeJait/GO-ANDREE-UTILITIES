package andretime

import "time"

type AndreTime interface {
	Now() time.Time
}

type (
	Time struct {
		time.Time
	}

	RealTime struct{}

	FakeTime struct {
		time time.Time
	}
)

func (t Time) UnixMilli() int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func NewRealTime() AndreTime {
	return &RealTime{}
}

func (rt *RealTime) Now() time.Time {
	return time.Now()
}

func NewFakeTime() AndreTime {
	return &FakeTime{
		time: time.Date(2020, time.August, 31, 0, 0, 0, 0, time.UTC),
	}
}

func NewFakeTimeAt(t time.Time) AndreTime {
	return &FakeTime{
		time: t,
	}
}

func (ft *FakeTime) Now() time.Time {
	return ft.time
}
