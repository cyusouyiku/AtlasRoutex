package utils

import "time"

func ParseDate(layout, value string) (time.Time, error) {
return time.Parse(layout, value)
}

func MustUTC(t time.Time) time.Time {
return t.UTC()
}
